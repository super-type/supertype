package dynamo

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/super-type/supertype/internal/reencryption"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
)

// GenerateSupertypeID generates a new Supertype ID for a given password
func GenerateSupertypeID(password string) (*string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"password": password,
	})
	if err != nil {
		return nil, storage.ErrMarshaling
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/generate-nuid-credentials", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, authenticating.ErrResponseBody
	}

	var supertypeID string
	json.Unmarshal([]byte(string(body)), &supertypeID)

	return &supertypeID, nil
}

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
// TODO should we make this more customizable and require passing in a map[string]*dynamodb.AttributeValue? or is that defeating the purpose
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
		return nil, storage.ErrFailedToReadDB
	}

	return result, nil
}

// CreateReencryptionKeys creates re-encryption keys for public keys returned from DynamoDB
// TODO should we figure out a way to abstract this and put it in our internal utils... although even passing in ScanOutput.Items is still DynamoDB-specific as Items is []map[string]*AttributeValue
func CreateReencryptionKeys(pkList *dynamodb.ScanOutput, skVendor *ecdsa.PrivateKey) (map[string][2]string, error) {
	connections := make(map[string][2]string)

	for i := 1; i < len(pkList.Items); i++ { // TODO not start at 1. Only doing now because entry 1 was a NuCypher pk
		pkTempStr := *(pkList.Items[i]["pk"].S)
		pkTemp, err := utils.StringToPublicKey(&pkTempStr)
		if err != nil {
			fmt.Printf("Error converting public key string to ECDSA Public Key: %v\n", err)
		}

		// Create re-encryption keys between each pairing
		rekey, pkX, err := reencryption.ReKeyGen(skVendor, &pkTemp)
		if err != nil {
			fmt.Printf("Error generating re-encryption key: %v\n", err) // TODO better error
		}

		rekeyStr := rekey.String()
		pkXStr := utils.PublicKeyToString(pkX)

		connections[pkTempStr] = [2]string{rekeyStr, pkXStr}
	}

	return connections, nil
}

// EstablishInitialConnections creates re-encryption keys between a newly-created vendor and all existing vendors
// TODO we will allow more granular access controls on vendor creation as we continue
func EstablishInitialConnections(svc *dynamodb.DynamoDB, pkVendor *string) (dynamodb.ScanOutput, error) {
	// Get all vendors' pks, except the vendor's own
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#pk": aws.String("pk"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(*pkVendor),
			},
		},
		FilterExpression:     aws.String("pk <> :pk"),
		ProjectionExpression: aws.String("#pk"),
		TableName:            aws.String("vendor"),
	}

	result, err := svc.Scan(input)
	if err != nil {
		fmt.Printf("Err scanning vendor table: %v\n", err)
	}

	return *result, nil
}
