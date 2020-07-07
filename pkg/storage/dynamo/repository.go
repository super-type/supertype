package dynamo

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

// CreateVendor creates a new vendor and adds it to DynamoDB
func (d *Storage) CreateVendor(vendor authenticating.Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error) {
	// TODO call a lambda to generate both an encrypting key pair and a verifying key pair
	// Initialize AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
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
		return nil, err
	}

	v := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &v)
	if err != nil {
		return nil, err
	}

	if v.Username != "" {
		fmt.Printf("Vendor %v already exists\n", v.Username)
		return nil, authenticating.ErrVendorAlreadyExists
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		fmt.Printf("Error generating keys: %v\n", err)
		return nil, err
	}

	// Generate Supertype ID
	requestBody, err := json.Marshal(map[string]string{
		"password": vendor.Password,
	})
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return nil, err
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/generate-nuid-credentials", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("API Error: %v\n", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
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
		fmt.Printf("Error marshalling vendor: %v\n", createVendor)
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Printf("Error adding item to DynamoDB: %v\n", err)
		return nil, err
	}

	fmt.Printf("Successfully added vendor %v, with SupertypeID %v\n", vendor.PublicKey, vendor.SupertypeID)
	keyPair := map[*ecdsa.PublicKey]*ecdsa.PrivateKey{pkVendor: skVendor}

	return keyPair, nil
}

// LoginVendor logs in the given vendor to the repository
func (d *Storage) LoginVendor(v authenticating.Vendor) error {
	// Initialize AWS Session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	tableName := "vendor"
	username := v.Username

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	})
	if err != nil {
		return err
	}

	vendor := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		return err
	}

	if vendor.Username == "" {
		return authenticating.ErrVendorNotFound
	}

	// TODO remove check username from lambda function
	// TODO create and call a lambda function to verify vendor using NuID

	return nil
}
