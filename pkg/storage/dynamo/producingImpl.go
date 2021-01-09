package dynamo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/producing"
	"github.com/super-type/supertype/pkg/storage"
	"go.uber.org/zap"
)

// Produce produces encyrpted data to Supertype
func (d *Storage) Produce(o producing.ObservationRequest, apiKey string) error {
	zap.S().Info("Producing new observation...")
	apiKeyHash := utils.GetAPIKeyHash(apiKey)
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "apiKeyHash", apiKeyHash)
	if err != nil || databaseAPIKeyHash == nil {
		zap.S().Errorf("Error getting API Key hash %s : %v", apiKeyHash, err)
		return err
	}

	pk, err := ScanDynamoDBWithKeyCondition("vendor", "pk", "apiKeyHash", apiKeyHash)
	if err != nil || pk == nil {
		zap.S().Errorf("Error getting vendor's public key given API Key: %v", err)
		return err
	}

	// Compare requesting API Key with our internal API Key. If they don't match, it's not coming from the vendor
	if *databaseAPIKeyHash != apiKeyHash {
		zap.S().Errorf("Stored API hash does not match given (%s) - potential malicious attempt!", apiKeyHash)
		return storage.ErrAPIKeyDoesNotMatch
	}

	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get current time
	currentTime := time.Now()

	// Create an observation to upload to DynamoDB
	d.Observation = Observation{
		Ciphertext:  o.Ciphertext + "|" + o.IV + "|" + o.Attribute,
		DateAdded:   currentTime.Format("2006-01-02 15:04:05.000000000"),
		PublicKey:   *pk,
		SupertypeID: o.SupertypeID,
	}

	// Upload new observation to DynamoDB
	err = PutItemInDynamoDB(d.Observation, o.Attribute, svc)
	if err != nil {
		zap.S().Errorf("Error putting vendor %s in DynamoDB: %v", fmt.Sprint(d.Observation), err)
		return err
	}

	// 1. Associate supertypeID with user, and get all vendors associated with that user
	username, err := ScanDynamoDBWithKeyCondition("user", "username", "supertypeID", o.SupertypeID)
	if err != nil {
		zap.S().Errorf("Username already exists: %v", err)
		return err
	}
	result, err := GetItemDynamoDB(svc, "user", "username", *username)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", *username, err)
		return err
	}
	user := authenticating.UserWithVendors{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		zap.S().Errorf("Error unmarshaling user: %v", err)
		return storage.ErrUnmarshaling
	}
	var webhooks []string

	// 2. Get all URLs associated with those vendors associated with the given user
	for _, vendor := range user.Vendors {
		username, err := ScanDynamoDBWithKeyCondition("vendor", "username", "pk", vendor)
		if err != nil {
			zap.S().Errorf("Username already exists: %v", err)
			return err
		}
		result, err := GetItemDynamoDB(svc, "vendor", "username", *username)
		if err != nil {
			zap.S().Errorf("Failed to get username %s from DynamoDB: %v", *username, err)
			return err
		}
		for _, url := range result.Item["webhooks"].L {
			webhooks = append(webhooks, *url.S)
		}
	}

	// 3. Iterate through all URLs for the published attribute (like all URLs for master-bedroom/lights/status)
	var webhookURLs []string
	destination := strings.Split(o.Attribute, "/")

	// Get attribute from subscribers
	result, err = GetItemDynamoDB(svc, "subscribers", "attribute", destination[0])
	if err != nil {
		zap.S().Errorf("Failed to get subscribers %v from DynamoDB: %v", destination[0], err)
		return err
	}
	var levels interface{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &levels)
	if err != nil {
		zap.S().Errorf("Error unmarshaling item %v", err)
		return err
	}

	if levels == nil {
		zap.S().Errorf("Attribute %v not found: %v", destination[0])
		return errors.New("Couldn't find attribute")
	}

	urls := GetSubscribersFromEndpoint(destination, levels)
	for _, url := range urls {
		webhookURLs = append(webhookURLs, url.(string))
	}

	// 4. If a URL matches the URLs associated with the vendors that are associated with a given user, send a Webhook POST request
	for _, webhookURL := range webhookURLs {
		if utils.Contains(webhooks, webhookURL) {
			color.Cyan("Sending POST to %v\n", webhookURL)

			requestBody, err := json.Marshal(map[string]string{
				"dateAdded":   currentTime.Format("2006-01-02 15:04:05.000000000"),
				"ciphertext":  d.Observation.Ciphertext,
				"pk":          d.Observation.PublicKey,
				"supertypeID": d.Observation.SupertypeID,
			})
			if err != nil {
				zap.S().Errorf("Error encoding request %s: %v", fmt.Sprint(requestBody), err)
				return err
			}

			client := &http.Client{}
			req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(requestBody))
			if err != nil {
				zap.S().Errorf("Error creating HTTP request: %v", err)
				return err
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("X-Supertype-Signature", apiKeyHash) // Send API Key Hash as signature, so vendors can verify it's from us

			resp, err := client.Do(req)
			if err != nil {
				zap.S().Errorf("Error sending HTTP request: %v", err)
				return err
			}

			// TODO make this more granular
			if resp.StatusCode >= 400 && resp.StatusCode < 600 {
				zap.S().Errorf("Invalid Webhook response: %v", resp.StatusCode)
				return errors.New("Error sending Webhook request")
			}
		}
	}

	zap.S().Info("Successfully produced new observation!")
	return nil
}
