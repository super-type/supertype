package dynamo

import (
	"fmt"

	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/storage"
	"go.uber.org/zap"
)

// Consume returns all observations at the requested attribute for the specified Supertype entity
func (d *Storage) Consume(c consuming.ObservationRequest, apiKey string) (*consuming.ObservationResponse, error) {
	zap.S().Info("Consuming attribute...")
	apiKeyHash := utils.GetAPIKeyHash(apiKey)
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "apiKeyHash", apiKeyHash)
	if err != nil {
		zap.S().Errorf("Error getting API Key hash %s :%v", apiKeyHash, err)
		return nil, err
	}

	if databaseAPIKeyHash == nil {
		zap.S().Errorf("API Key hash %s is nil", apiKeyHash)
	}

	// Compare requesting API Key with our internal API Key. If they don't match, it's not coming from the vendor
	if *databaseAPIKeyHash != apiKeyHash {
		zap.S().Errorf("Stored API hash does not match given (%s) - potential malicious attempt!", apiKeyHash)
		return nil, storage.ErrAPIKeyDoesNotMatch
	}

	// Initialize AWS session
	svc := utils.SetupAWSSession()

	val, err := GetItemDynamoDB(svc, c.Attribute, "supertypeID", c.SupertypeID)
	if err != nil {
		zap.S().Errorf("Failed to get username %s from DynamoDB: %v", c.SupertypeID, err)
		return nil, err
	}

	observation := consuming.ObservationResponse{
		Ciphertext:  *(val.Item["ciphertext"].S),
		DateAdded:   *(val.Item["dateAdded"].S),
		PublicKey:   *(val.Item["pk"].S),
		SupertypeID: *(val.Item["supertypeID"].S),
	}

	zap.S().Infof("Successfully retrieved observation %s", fmt.Sprint(observation))
	return &observation, nil
}
