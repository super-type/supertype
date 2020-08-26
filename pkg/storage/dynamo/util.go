package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatih/color"
	"github.com/super-type/supertype/pkg/storage"
)

// SetupAWSSession starts an AWS session
func SetupAWSSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	return svc
}

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
