package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
)

// GetItemDynamoDB gets an item from DynamoDB
func GetItemDynamoDB(svc *dynamodb.DynamoDB, tableName string, attribute string, value string) (*dynamodb.GetItemOutput, error) {
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

// ScanDynamoDB gets given attribute
func ScanDynamoDB(table string, attribute string, value string) (*dynamodb.ScanOutput, error) {
	svc := utils.SetupAWSSession()

	filt := expression.Name(attribute).Contains(value)

	proj := expression.NamesList(
		expression.Name(attribute),
	)

	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		color.Red("Error building expression", err)
		return nil, err
	}

	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(table),
	}

	result, err := svc.Scan(scanInput)
	if err != nil {
		color.Red("Error scanning", err)
		return nil, err
	}

	return result, nil
}

// ScanDynamoDBWithKeyCondition gets given attribute with specific key condition
func ScanDynamoDBWithKeyCondition(table string, attribute string, keyCondition string, keyConditionValue string) (*string, error) {
	svc := utils.SetupAWSSession()

	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#" + attribute: aws.String(attribute),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":" + keyCondition: {
				S: aws.String(keyConditionValue),
			},
		},
		FilterExpression:     aws.String(keyCondition + " = :" + keyCondition),
		ProjectionExpression: aws.String("#" + attribute),
		TableName:            aws.String(table),
	}

	result, err := svc.Scan(scanInput)
	if err != nil || CheckAWSScanChain(result, attribute) {
		return nil, err
	}

	return result.Items[0][attribute].S, nil
}

// CheckAWSScanChain checks all items through an AWS DynamoDB scan to make sure none are nil
func CheckAWSScanChain(so *dynamodb.ScanOutput, attribute string) bool {
	if so.Items == nil || len(so.Items) == 0 || so.Items[0][attribute] == nil || so.Items[0][attribute].S == nil {
		return true
	}
	return false
}

// PutItemInDynamoDB adds an item to DynamoDb
func PutItemInDynamoDB(in interface{}, table string, svc *dynamodb.DynamoDB) error {
	// Upload new vendor to DynamoDB
	av, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		color.Red("Error marshaling data")
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		color.Red("Failed to write to database")
		return err
	}

	return nil
}
