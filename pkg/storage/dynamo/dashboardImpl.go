package dynamo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/storage"
)

// ListAttributes returns all observations in the Supertype ecosystem
func (d *Storage) ListAttributes() ([]string, error) {
	zap.S().Info("Listing attributes...")
	// Initialize AWS session
	svc := utils.SetupAWSSession()
	input := &dynamodb.ListTablesInput{}

	var response []string

	for {
		// Get list of tables
		result, err := svc.ListTables(input)
		if err != nil {
			if _, ok := err.(awserr.Error); ok {
				zap.S().Error("Dynamo internal server error")
				return nil, dashboard.ErrDynamoInternalError
			}
			zap.S().Errorf("Dynamo error: %v", err)
			return nil, dashboard.ErrDynamoError
		}

		for _, n := range result.TableNames {
			if *n != "poc-todo" && *n != "public-keys" && *n != "vendor" {
				response = append(response, *n)
			}
		}

		// assign the last read tablename as the start for our next call to the ListTables function
		// the maximum number of table names returned in a call is 100 (default), which requires us to make
		// multiple calls to the ListTables function to retrieve all table names
		input.ExclusiveStartTableName = result.LastEvaluatedTableName

		if result.LastEvaluatedTableName == nil {
			break
		}
	}

	zap.S().Info("Successfully retrieved attributes")
	return response, nil
}

// RegisterWebhook creates a new webhook on a vendor's request
func (d *Storage) RegisterWebhook(webhookRequest dashboard.WebhookRequest, apiKey string) error {
	zap.S().Info("Registering webhook...")
	apiKeyHash := utils.GetAPIKeyHash(apiKey)
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "apiKeyHash", apiKeyHash)
	if err != nil || databaseAPIKeyHash == nil {
		zap.S().Errorf("Error getting vendor's public key given API Key: %v", err)
		return err
	}

	// Compare requesting API Key with our internal API Key. If they don't match, it's not coming from the vendor
	if *databaseAPIKeyHash != apiKeyHash {
		zap.S().Errorf("Stored API hash does not match given (%s) - potential malicious attempt!", apiKeyHash)
		return storage.ErrAPIKeyDoesNotMatch
	}

	// Parse endpoint, assuming it was validated on client side (or, it'll just throw an error if it's wrong)
	endpoint := strings.Split(webhookRequest.Endpoint, "/")
	breakpoint := 0
	for i := 0; i < len(endpoint); i++ {
		breakpoint = i
		// Everything after the /supertype/ qualifier is our Supertype attribute
		if endpoint[i] == "supertype" {
			breakpoint++
			break
		}
	}
	destination := endpoint[breakpoint:]

	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get attribute from subscribers
	result, err := GetItemDynamoDB(svc, "subscribers", "attribute", destination[0])
	if err != nil {
		zap.S().Errorf("Failed to get subscribers to %v from DynamoDB: %v", destination[0], err)
		return err
	}
	var levels interface{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &levels)
	if err != nil {
		zap.S().Errorf("Error unmarshaling levels: %v", err)
		return err
	}

	switch destination[0] {
	case "master-bedroom":
		var attribute dashboard.MasterBedroom
		err = dynamodbattribute.UnmarshalMap(result.Item, &attribute)
		if err != nil {
			zap.S().Errorf("Error unmarshaling attribute: %v", err)
			return err
		}

		b, err := json.Marshal(attribute)
		if err != nil {
			zap.S().Errorf("Error encoding request: %v", err)
			return err
		}

		err = utils.ValidateNewSubscriberURL(string(b), webhookRequest.Endpoint)
		if err != nil {
			zap.S().Errorf("Error validating endpoint %v : %v", webhookRequest.Endpoint, err)
			return err
		}

		urls := GetSubscribersFromEndpoint(destination, levels)
		updatedAttribute, err := utils.AppendToSubscribers(string(b), urls, webhookRequest.Endpoint)
		if err != nil {
			zap.S().Errorf("Error appending to endpoint: %v", err)
			return err
		}

		resp := dashboard.MasterBedroom{}
		err = json.Unmarshal([]byte(*updatedAttribute), &resp)
		if err != nil {
			zap.S().Errorf("Error unmarshaling attribute %s : %v", *updatedAttribute, err)
			return err
		}

		err = PutItemInDynamoDB(resp, "subscribers", svc)
		if err != nil {
			zap.S().Errorf("Error putting item %s in DynamoDB: %v", fmt.Sprint(resp), err)
			return err
		}
	// This could get ugly, fast... we should try to think of a way to do this more programmatically
	case "living-room":
	case "laundry-room":
	case "kitchen":
	case "kids-bedroom":
	case "guest-bedroom":
	case "garage":
	case "bathroom":
	default:
		zap.S().Errorf("Invalid attribute %s : %v", destination[0])
		return errors.New("Invalid attribute")
	}

	username, err := ScanDynamoDBWithKeyCondition("vendor", "username", "apiKeyHash", apiKeyHash)
	if err != nil {
		zap.S().Errorf("Username already exists: %v", err)
		return err
	}
	result, err = GetItemDynamoDB(svc, "vendor", "username", *username)
	if err != nil {
		zap.S().Errorf("Failed to get vendor from DynamoDB: %v", err)
		return err
	}

	vendor := authenticating.CreateVendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		zap.S().Errorf("Error unmarshaling vendor: %v", err)
		return storage.ErrUnmarshaling
	}

	updatedWebhooks := append(vendor.Webhooks, webhookRequest.Endpoint)

	updatedVendor := authenticating.CreateVendor{
		FirstName:      vendor.FirstName,
		LastName:       vendor.LastName,
		Email:          vendor.Email,
		BusinessName:   vendor.BusinessName,
		Username:       vendor.Username,
		PublicKey:      vendor.PublicKey,
		APIKeyHash:     apiKeyHash,
		SupertypeID:    vendor.SupertypeID,
		AccountBalance: vendor.AccountBalance,
		Webhooks:       updatedWebhooks,
	}

	// TODO this needs to be an atomic transaction. If we can't add the URL to vendor, we need to remove it from "subscribers"
	// The URL can't exist in one place but not the other... maybe this is bad practice without a single source of truth?
	err = PutItemInDynamoDB(updatedVendor, "vendor", svc)
	if err != nil {
		zap.S().Errorf("Error putting vendor %s in DynamoDB: %v", fmt.Sprint(updatedVendor), err)
		return err
	}

	zap.S().Info("Successfully registered webhook!")
	return nil
}
