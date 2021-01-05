package dynamo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/storage"
)

// ListAttributes returns all observations in the Supertype ecosystem
func (d *Storage) ListAttributes() ([]string, error) {
	// Initialize AWS session
	svc := utils.SetupAWSSession()
	input := &dynamodb.ListTablesInput{}

	var response []string

	for {
		// Get list of tables
		result, err := svc.ListTables(input)
		if err != nil {
			if _, ok := err.(awserr.Error); ok {
				color.Red("Dynamo internal server error")
				return nil, dashboard.ErrDynamoInternalError
			}
			color.Red("Dynamo error")
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

	return response, nil
}

// RegisterWebhook creates a new webhook on a vendor's request
func (d *Storage) RegisterWebhook(webhookRequest dashboard.WebhookRequest, apiKey string) error {
	apiKeyHash := utils.GetAPIKeyHash(apiKey)
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "apiKeyHash", apiKeyHash)
	if err != nil || databaseAPIKeyHash == nil {
		return err
	}

	// Compare requesting API Key with our internal API Key. If they don't match, it's not coming from the vendor
	if *databaseAPIKeyHash != apiKeyHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
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
		return err
	}
	var levels interface{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &levels)
	if err != nil {
		return err
	}

	switch destination[0] {
	case "master-bedroom":
		var attribute dashboard.MasterBedroom
		err = dynamodbattribute.UnmarshalMap(result.Item, &attribute)
		if err != nil {
			return err
		}

		b, err := json.Marshal(attribute)
		if err != nil {
			return err
		}

		err = utils.ValidateNewSubscriberURL(string(b), webhookRequest.Endpoint)
		if err != nil {
			return err
		}

		urls := GetSubscribersFromEndpoint(destination, levels)
		updatedAttribute, err := utils.AppendToSubscribers(string(b), urls, webhookRequest.Endpoint)
		if err != nil {
			return err
		}

		resp := dashboard.MasterBedroom{}
		err = json.Unmarshal([]byte(*updatedAttribute), &resp)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = PutItemInDynamoDB(resp, "subscribers", svc)
		if err != nil {
			fmt.Println(err)
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
		return errors.New("Invalid attribute")
	}

	username, err := ScanDynamoDBWithKeyCondition("vendor", "username", "apiKeyHash", apiKeyHash)
	if err != nil {
		return err
	}
	result, err = GetItemDynamoDB(svc, "vendor", "username", *username)
	if err != nil {
		return err
	}

	vendor := authenticating.CreateVendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		color.Red("Error unmarshaling data")
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
		fmt.Println(err)
		return err
	}

	return nil
}
