package dynamo

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fatih/color"
	"github.com/super-type/supertype/pkg/producing"
	"github.com/super-type/supertype/pkg/storage"
)

// Produce produces encyrpted data to Supertype
func (d *Storage) Produce(o producing.ObservationRequest) error {
	// Get the skHash of the given vendor
	skHash, err := GetSkHash(o.PublicKey)

	// Compare requesting skHash with our internal skHash. If they don't match, it's not coming from the vendor
	if *skHash != o.SkHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
		return storage.ErrSkHashDoesNotMatch
	}

	// Initialize AWS session
	svc := SetupAWSSession()

	// Get current time
	currentTime := time.Now()

	// Create an observation to upload to DynamoDB
	d.Observation = Observation{
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
	av, err := dynamodbattribute.MarshalMap(d.Observation)
	if err != nil {
		color.Red("Error marshaling data")
		return storage.ErrMarshaling
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: &o.Attribute,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		color.Red("Failed to write to database")
		return storage.ErrFailedToWriteDB
	}

	return nil
}

// GetVendorComparisonMetadata returns lists of both all vendors, and all of the requesting vendors' connections
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
