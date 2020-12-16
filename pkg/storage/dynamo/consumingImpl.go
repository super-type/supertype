package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/storage"
)

// Consume returns all observations at the requested attribute for the specified Supertype entity
func (d *Storage) Consume(c consuming.ObservationRequest, apiKeyHash string) (*[]consuming.ObservationResponse, error) {
	databaseAPIKeyHash, err := ScanDynamoDBWithKeyCondition("vendor", "apiKeyHash", "pk", c.PublicKey)
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

	// Get all observations for the specified attribute with user's Supertype ID
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#ciphertext":  aws.String("ciphertext"),
			"#dateAdded":   aws.String("dateAdded"),
			"#pk":          aws.String("pk"),          // ? do we need this
			"#supertypeID": aws.String("supertypeID"), // ? do we need this
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":supertypeID": {
				S: aws.String(c.SupertypeID),
			},
		},
		FilterExpression:     aws.String("supertypeID = :supertypeID"),
		ProjectionExpression: aws.String("#ciphertext, #dateAdded, #pk, #supertypeID"),
		TableName:            aws.String(c.Attribute),
	}

	result, err := svc.Scan(input)
	if err != nil {
		return nil, err
	}

	// Get observations from result
	observations := make([]consuming.ObservationResponse, 0)
	for _, observation := range result.Items {
		tempObservation := consuming.ObservationResponse{
			Ciphertext:  *(observation["ciphertext"].S),
			DateAdded:   *(observation["dateAdded"].S),
			PublicKey:   *(observation["pk"].S),
			SupertypeID: *(observation["supertypeID"].S),
		}
		observations = append(observations, tempObservation)
	}

	return &observations, nil
}
