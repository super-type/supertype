package dynamo

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/super-type/supertype/pkg/storage"

	"github.com/super-type/supertype/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/pkg/authenticating"
)

// Storage keeps data in dynamo
type Storage struct {
	vendor Vendor
}

func setupAWSSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	return svc
}

func listAllVendors(svc *dynamodb.DynamoDB) {
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#pk": aws.String("pk"),
		},
		ProjectionExpression: aws.String("#pk"),
		TableName:            aws.String("vendor"),
	}

	result, err := svc.Scan(input)
	if err != nil {
		fmt.Printf("Err scanning vendor table: %v\n", err)
	}

	fmt.Printf("result: %v\n", result)
}

// CreateVendor creates a new vendor and adds it to DynamoDB
func (d *Storage) CreateVendor(vendor authenticating.Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error) {
	// Initialize AWS session
	svc := setupAWSSession()
	tableName := "vendor"

	// Check username doesn't exist
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(vendor.Username),
			},
		},
	})
	if err != nil {
		return nil, storage.ErrFailedToReadDB
	}

	v := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &v)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	if v.Username != "" {
		return nil, authenticating.ErrVendorAlreadyExists
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		return nil, keys.ErrFailedToGenerateKeys
	}

	// Generate Supertype ID
	requestBody, err := json.Marshal(map[string]string{
		"password": vendor.Password,
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

	// Finalize attributes for new Supertype user
	vendor.PublicKey = utils.PublicKeyToString(pkVendor)
	vendor.SupertypeID = supertypeID
	vendor.Connections = make(map[string]string)

	createVendor := authenticating.CreateVendor{
		FirstName:   vendor.FirstName,
		LastName:    vendor.LastName,
		Username:    vendor.Username,
		PublicKey:   utils.PublicKeyToString(pkVendor),
		SupertypeID: supertypeID,
		Connections: make(map[string]string),
	}

	// Upload new vendor to DynamoDB
	av, err := dynamodbattribute.MarshalMap(createVendor)
	if err != nil {
		return nil, storage.ErrMarshaling
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return nil, storage.ErrFailedToWriteDB
	}

	keyPair := map[*ecdsa.PublicKey]*ecdsa.PrivateKey{pkVendor: skVendor}

	// TODO re-encrypt the new vendor between its public key and all other vendors' public keys
	listAllVendors(svc)

	return keyPair, nil
}

// LoginVendor logs in the given vendor to the repository
func (d *Storage) LoginVendor(v authenticating.Vendor) (*authenticating.AuthenticatedVendor, error) {
	// Initialize AWS Session
	svc := setupAWSSession()
	tableName := "vendor"

	// Check vendor exists and get object
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(v.Username),
			},
		},
	})
	if err != nil {
		return nil, storage.ErrFailedToReadDB
	}

	vendor := authenticating.AuthenticatedVendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	if vendor.Username == "" {
		return nil, authenticating.ErrVendorNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		// TODO we need a better way to do this
		"password":    v.Password,
		"supertypeID": vendor.SupertypeID,
	})
	if err != nil {
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, authenticating.ErrResponseBody
	}

	var jwt *string
	json.Unmarshal([]byte(string(body)), &jwt)
	vendor.JWT = *jwt

	return &vendor, nil
}
