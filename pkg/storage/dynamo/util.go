package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/storage"
)

// GetFromDynamoDB gets an item from DynamoDB
// NOTE this should stay in utils becuase while we're currently only using it for authenticating, it may be more prevalent
func GetFromDynamoDB(svc *dynamodb.DynamoDB, tableName string, attribute string, value string) (*dynamodb.GetItemOutput, error) {
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			attribute: {
				S: aws.String(value),
			},
		},
	})
	if err != nil {
		color.Red("Failed to read from database")
		return nil, storage.ErrFailedToReadDB
	}

	return result, nil
}

// GetSkHash gets the secret key hash of the given vendor
func GetSkHash(pk string) (*string, error) {
	svc := utils.SetupAWSSession()

	// Get the skHash of the given vendor
	skHashInput := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#skHash": aws.String("skHash"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(pk),
			},
		},
		FilterExpression:     aws.String("pk = :pk"),
		ProjectionExpression: aws.String("#skHash"),
		TableName:            aws.String("vendor"),
	}

	skHash, err := svc.Scan(skHashInput)
	if err != nil || CheckAWSScanChain(skHash) {
		return nil, err
	}

	return skHash.Items[0]["skHash"].S, nil
}

// CheckAWSScanChain checks all items through an AWS DynamoDB scan to make sure none are nil
func CheckAWSScanChain(so *dynamodb.ScanOutput) bool {
	if so.Items == nil || so.Items[0]["skHash"] == nil || so.Items[0]["skHash"].S == nil {
		return true
	}
	return false
}
