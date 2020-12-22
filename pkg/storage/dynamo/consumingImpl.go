package dynamo

import (
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/storage"
)

// Consume returns all observations at the requested attribute for the specified Supertype entity
func (d *Storage) Consume(c consuming.ObservationRequest, apiKey string) (*consuming.ObservationResponse, error) {
	apiKeyHash := utils.GetAPIKeyHash(apiKey)
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "apiKeyHash", apiKeyHash)
	if err != nil || databaseAPIKeyHash == nil {
		return nil, err
	}

	// Compare requesting API Key with our internal API Key. If they don't match, it's not coming from the vendor
	if *databaseAPIKeyHash != apiKeyHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
		return nil, storage.ErrAPIKeyDoesNotMatch
	}

	// Initialize AWS session
	svc := utils.SetupAWSSession()

	val, err := GetItemDynamoDB(svc, c.Attribute, "supertypeID", c.SupertypeID)
	if err != nil {
		return nil, err
	}

	observation := consuming.ObservationResponse{
		Ciphertext:  *(val.Item["ciphertext"].S),
		DateAdded:   *(val.Item["dateAdded"].S),
		PublicKey:   *(val.Item["pk"].S),
		SupertypeID: *(val.Item["supertypeID"].S),
	}

	return &observation, nil
}
