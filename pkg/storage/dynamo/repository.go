package dynamo

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/super-type/supertype/pkg/consuming"

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
func (d *Storage) Produce(o producing.ObservationRequest) error {
	// Initialize AWS session
	svc := SetupAWSSession()

	// Get current time
	currentTime := time.Now()

	// Create an observation to upload to DynamoDB
	d.observation = Observation{
		Ciphertext: o.Ciphertext,
		// Capsule:     o.Capsule,
		CapsuleE:    o.CapsuleE,
		CapsuleV:    o.CapsuleV,
		CapsuleS:    o.CapsuleS,
		DateAdded:   currentTime.Format("2006-01-02 15:04:05.000000000"),
		PublicKey:   o.PublicKey,
		SupertypeID: o.SupertypeID,
	}

	// Upload new observation to DynamoDB
	av, err := dynamodbattribute.MarshalMap(d.observation)
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

// Consume returns all observations at the requested attribute for the specified Supertype entity
func (d *Storage) Consume(c consuming.ObservationRequest) (*[]consuming.ObservationResponse, error) {
	// Initialize AWS session
	svc := SetupAWSSession()

	// Get all observations for the specified attribute with user's Supertype ID
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#ciphertext": aws.String("ciphertext"),
			// "#capsule":     aws.String("capsule"),
			"#capsuleE":    aws.String("capsuleE"),
			"#capsuleV":    aws.String("capsuleV"),
			"#capsuleS":    aws.String("capsuleS"),
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
		ProjectionExpression: aws.String("#ciphertext, #capsuleE, #capsuleV, #capsuleS, #dateAdded, #pk, #supertypeID"),
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
			CapsuleE:    *(observation["capsuleE"].S),
			CapsuleV:    *(observation["capsuleV"].S),
			CapsuleS:    *(observation["capsuleS"].S),
			DateAdded:   *(observation["dateAdded"].S),
			PublicKey:   *(observation["pk"].S),
			SupertypeID: *(observation["supertypeID"].S),
		}
		observations = append(observations, tempObservation)
	}

	// Get the rekey, pkX for each observation and add it to the response (adds nothing if it's from the consuming vendor)
	for i, observation := range observations {
		if observation.PublicKey != c.PublicKey {
			input = &dynamodb.ScanInput{
				ExpressionAttributeNames: map[string]*string{
					"#connections": aws.String("connections"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":pk": {
						S: aws.String(observation.PublicKey),
					},
				},
				FilterExpression:     aws.String("pk = :pk"),
				ProjectionExpression: aws.String("#connections"),
				TableName:            aws.String("vendor"),
			}

			result, err := svc.Scan(input)
			if err != nil || result.Items == nil {
				return nil, err
			}

			// rekey, pkX for associated <pkObservation, pkVendor>
			connectionMetadata := result.Items[0]["connections"].M[c.PublicKey].L
			rekey := connectionMetadata[0].S
			pkX := connectionMetadata[1].S

			reencryptionMetadata := [2]string{*rekey, *pkX}
			observations[i].ReencryptionMetadata = reencryptionMetadata
		}
		// TODO put something in here...
	}

	return &observations, nil
}

// GetVendorComparisonMetadata returns lists of both all vendors, and all of the requesting vendors' connections
// TODO once we implement a system to replicate DynamoDB data in Elasticsearch, we should use Elasticsearch
func (d *Storage) GetVendorComparisonMetadata(o producing.ObservationRequest) (*producing.MetadataResponse, error) {
	// Initialize AWS session
	svc := SetupAWSSession()

	// Get all vendor public keys
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#pk": aws.String("pk"),
		},
		ProjectionExpression: aws.String("#pk"),
		TableName:            aws.String("vendor"),
	}

	vendors, err := svc.Scan(input)
	if err != nil {
		return nil, err
	}

	// Iterate through vendor PK AWS response
	var pkVendors []string
	for i := range vendors.Items {
		pkVendors = append(pkVendors, *vendors.Items[i]["pk"].S)
	}

	// Get connections of requesting vendor
	input = &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#connections": aws.String("connections"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(o.PublicKey),
			},
		},
		FilterExpression:     aws.String("pk = :pk"),
		ProjectionExpression: aws.String("#connections"),
		TableName:            aws.String("vendor"),
	}

	connections, err := svc.Scan(input)
	if err != nil {
		return nil, err
	}

	// Iterate through connections keys (i.e. public keys of vendors we have connections for)
	var pkConnections []string
	for k := range connections.Items[0]["connections"].M {
		pkConnections = append(pkConnections, k)
	}

	response := producing.MetadataResponse{
		VendorConnections: pkConnections,
		Vendors:           pkVendors,
	}

	return &response, nil
}

// AddReencryptionKeys adds newly-created re-encryption keys for pre-existing vendors
func (d *Storage) AddReencryptionKeys(r producing.ReencryptionKeysRequest) error {
	// Initialize AWS session
	svc := SetupAWSSession()

	// Get connections of requesting vendor and username, since as of 07/2020, DynamoDB doesn't support updates on GSIs
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#connections": aws.String("connections"),
			"#username":    aws.String("username"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(r.PublicKey),
			},
		},
		FilterExpression:     aws.String("pk = :pk"),
		ProjectionExpression: aws.String("#connections, #username"),
		TableName:            aws.String("vendor"),
	}

	connectionsResp, err := svc.Scan(input)
	if err != nil {
		return err
	}

	username := connectionsResp.Items[0]["username"].S
	connections := connectionsResp.Items[0]["connections"].M

	// Add each new connection to the map
	for k, v := range r.Connections {
		dynamoV := make([]*dynamodb.AttributeValue, 2)
		dynamoV[0] = &dynamodb.AttributeValue{
			S: aws.String(v[0]),
		}
		dynamoV[1] = &dynamodb.AttributeValue{
			S: aws.String(v[1]),
		}

		connectionsResp.Items[0]["connections"].M[k] = &dynamodb.AttributeValue{
			L: []*dynamodb.AttributeValue{
				{
					S: aws.String(v[0]),
				},
				{
					S: aws.String(v[1]),
				},
			},
		}
	}

	// Update connections in DynamoDB
	updateInput := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#connections": aws.String("connections"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":connections": {
				M: connections,
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: username,
			},
		},
		TableName:        aws.String("vendor"),
		UpdateExpression: aws.String("SET #connections = :connections"),
	}

	_, err = svc.UpdateItem(updateInput)
	if err != nil {
		return err
	}

	return nil
}
