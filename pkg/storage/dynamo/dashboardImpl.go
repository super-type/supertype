package dynamo

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/storage"
)

// ListObservations returns all observations in the Supertype ecosystem
func (d *Storage) ListObservations(o dashboard.ObservationRequest, apiKeyHash string) ([]string, error) {
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
