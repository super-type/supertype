package dynamo

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/super-type/supertype/pkg/producing"

	"github.com/super-type/supertype/pkg/storage"

	"github.com/super-type/supertype/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/pkg/authenticating"
)

// Storage keeps data in dynamo
type Storage struct {
	vendor      Vendor
	observation Observation
}

var tableName = "vendor"

// CreateVendor creates a new vendor and adds it to DynamoDB
func (d *Storage) CreateVendor(v authenticating.Vendor) (*[2]string, error) {
	// Initialize AWS session
	svc := SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetFromDynamoDB(svc, tableName, "username", v.Username)
	if err != nil {
		return nil, err
	}
	vendor := Vendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	// Check username doesn't exist
	if vendor.Username != "" {
		return nil, authenticating.ErrVendorAlreadyExists
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		return nil, keys.ErrFailedToGenerateKeys
	}

	// Generate Supertype ID
	supertypeID, err := GenerateSupertypeID(v.Password)
	if err != nil {
		return nil, err
	}

	// Generate re-encryption keys
	pk := utils.PublicKeyToString(pkVendor)
	pkList, err := EstablishInitialConnections(svc, &pk)
	if err != nil {
		return nil, storage.ErrGetListPublicKeys
	}
	rekeys, err := CreateReencryptionKeys(pkList, skVendor)
	if err != nil {
		return nil, keys.ErrFailedToGenerateReencryptionKeys
	}

	// Create a final vendor with which to upload
	createVendor := authenticating.CreateVendor{
		FirstName:      v.FirstName,
		LastName:       v.LastName,
		Email:          v.Email,
		BusinessName:   v.BusinessName,
		Username:       v.Username,
		PublicKey:      utils.PublicKeyToString(pkVendor),
		SupertypeID:    *supertypeID,
		Connections:    rekeys,
		AccountBalance: 0.0,
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

	// TODO change this back in case we need it. Maybe storing the D value is fine?
	// keyPair := [2]string{utils.PublicKeyToString(pkVendor), utils.PrivateKeyToString(skVendor)}
	// todo research if this is safe enough to have user "store" as their secret key... stored offline of course
	keyPair := [2]string{utils.PublicKeyToString(pkVendor), skVendor.D.String()}

	return &keyPair, nil
}

// LoginVendor logs in the given vendor to the repository
func (d *Storage) LoginVendor(v authenticating.Vendor) (*authenticating.AuthenticatedVendor, error) {
	// Initialize AWS Session
	svc := SetupAWSSession()

	// Get username from DynamoDB
	result, err := GetFromDynamoDB(svc, tableName, "username", v.Username)
	if err != nil {
		return nil, err
	}
	vendor := authenticating.AuthenticatedVendor{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &vendor)
	if err != nil {
		return nil, storage.ErrUnmarshaling
	}

	// Check vendor exists and get object
	if vendor.Username == "" {
		return nil, authenticating.ErrVendorNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
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

// Produce produces encyrpted data to Supertype
func (d *Storage) Produce(o producing.Observation) error {
	// Initialize AWS session
	svc := SetupAWSSession()

	// Get current time
	currentTime := time.Now()

	// Create an observation to upload to DynamoDB
	observation := Observation{
		Ciphertext: o.Ciphertext,
		Capsule:    o.Capsule,
		DateAdded:  currentTime.Format("2006-01-02 15:04:05.000000000"),
	}

	// Upload new observation to DynamoDB
	av, err := dynamodbattribute.MarshalMap(observation)
	if err != nil {
		return storage.ErrMarshaling
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: &o.Attribute,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return storage.ErrFailedToWriteDB
	}

	return nil
}
