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
)

// Produce produces encyrpted data to Supertype
func (d *Storage) Produce(o producing.ObservationRequest, apiKey string) error {
	apiKeyHash := utils.GetAPIKeyHash(apiKey)
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "apiKeyHash", apiKeyHash)
	if err != nil || databaseAPIKeyHash == nil {
		return err
	}

	pk, err := ScanDynamoDBWithKeyCondition("vendor", "pk", "apiKeyHash", apiKeyHash)
	if err != nil || pk == nil {
		fmt.Println(err)
		return err
	}

	// Compare requesting API Key with our internal API Key. If they don't match, it's not coming from the vendor
	if *databaseAPIKeyHash != apiKeyHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
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
		return err
	}

	// TODO after uploading, we want to
	// 1. Associate supertypeID with user, and get all vendors associated with that user
	username, err := ScanDynamoDBWithKeyCondition("user", "username", "supertypeID", o.SupertypeID)
	if err != nil {
		return err
	}
	result, err := GetItemDynamoDB(svc, "user", "username", *username)
	if err != nil {
		return err
	}
	user := authenticating.UserWithVendors{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		color.Red("Error unmarshaling data")
		return storage.ErrUnmarshaling
	}
	var webhooks []string

	// 2. Get all URLs associated with those vendors associated with the given user
	for _, vendor := range user.Vendors {
		username, err := ScanDynamoDBWithKeyCondition("vendor", "username", "pk", vendor)
		if err != nil {
			return err
		}
		result, err := GetItemDynamoDB(svc, "vendor", "username", *username)
		if err != nil {
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
		return err
	}
	var levels interface{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &levels)
	if err != nil {
		return err
	}

	if levels == nil {
		return errors.New("Couldn't find attribute")
	}

	for i := 0; i < len(destination); i++ {
		if levels.(map[string]interface{})[destination[i]] == nil {
			continue
		}
		levels = levels.(map[string]interface{})[destination[i]]
	}

	levels = levels.(map[string]interface{})["subscribers"]
	urls := levels.([]interface{})

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
				color.Red("Error marshaling data")
				return err
			}

			client := &http.Client{}
			req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(requestBody))
			if err != nil {
				fmt.Println(err)
				return err
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("X-Supertype-Key", apiKeyHash) // Send API Key Hash as signature, so vendors can verify it's from us

			resp, err := client.Do(req)
			if err != nil {
				color.Red("Error sending Webhook request")
				return err
			}

			// TODO make this more granular
			if resp.StatusCode >= 400 && resp.StatusCode < 600 {
				return errors.New("Error sending Webhook request")
			}
		}
	}

	return nil
}
