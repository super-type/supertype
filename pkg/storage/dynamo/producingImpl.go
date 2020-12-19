package dynamo

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
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

	return nil
}
