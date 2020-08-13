package dynamo

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/keys"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/storage"
)

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
		color.Red("Vendor already exists")
		return nil, authenticating.ErrVendorAlreadyExists
	}

	// Generate key pair for new vendor
	skVendor, pkVendor, err := keys.GenerateKeys()
	if err != nil {
		color.Red("Failed to generate keys")
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
		color.Red("Failed to get list of public keys")
		return nil, storage.ErrGetListPublicKeys
	}
	rekeys, err := CreateReencryptionKeys(pkList, skVendor)
	if err != nil {
		color.Red("Failed to generate re-encryption keys")
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
		color.Red("Error marshaling data")
		return nil, storage.ErrMarshaling
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		color.Red("Failed to write to database")
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
		color.Red("Error unmarshaling data")
		return nil, storage.ErrUnmarshaling
	}

	// Check vendor exists and get object
	if vendor.Username == "" {
		color.Red("Vendor not found")
		return nil, authenticating.ErrVendorNotFound
	}

	// Authenticate with NuID
	requestBody, err := json.Marshal(map[string]string{
		"password":    v.Password,
		"supertypeID": vendor.SupertypeID,
	})
	if err != nil {
		color.Red("Error encoding data")
		return nil, storage.ErrEncoding
	}

	resp, err := http.Post("https://z1lwetrbfe.execute-api.us-east-1.amazonaws.com/default/login-vendor", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		color.Red("Error requesting Supertype API")
		return nil, authenticating.ErrRequestingAPI
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		color.Red("API request gave bad response status")
		return nil, authenticating.ErrRequestingAPI
	}

	jwt, err := authenticating.GenerateJWT(vendor.Username)
	if err != nil {
		color.Red("Could not generate JWT")
		return nil, authenticating.ErrRequestingAPI
	}
	vendor.JWT = *jwt

	return &vendor, nil
}
