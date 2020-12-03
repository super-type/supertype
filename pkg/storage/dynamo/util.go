package dynamo

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
)

// GetFromDynamoDB gets an item from DynamoDB
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
		color.Red("Failed to read from database: ", err)
		return nil, err
	}

	return result, nil
}

// GetSkHash gets the secret key hash of the given vendor
// TODO standardize this with GetEmail
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

// GetEmail gets the email of the given vendor
// TODO standardize this with GetSkHash
func GetEmail(email string) (*dynamodb.ScanOutput, error) {
	svc := utils.SetupAWSSession()

	filt := expression.Name("email").Contains(email)

	proj := expression.NamesList(
		expression.Name("email"),
	)

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		fmt.Println(err)
	}

	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String("vendor"),
	}

	result, err := svc.Scan(scanInput)
	if err != nil {
		color.Red("Error scanning")
		return nil, err
	}

	return result, nil
}

// CheckAWSScanChain checks all items through an AWS DynamoDB scan to make sure none are nil
func CheckAWSScanChain(so *dynamodb.ScanOutput) bool {
	// TODO do we need more values than just nil and initial empty check?
	if so.Items == nil || len(so.Items) == 0 || so.Items[0]["email"] == nil || so.Items[0]["email"].S == nil {
		return true
	}
	return false
}
