package dynamo

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/producing"
	"github.com/super-type/supertype/pkg/storage"
)

// Produce produces encyrpted data to Supertype
func (d *Storage) Produce(o producing.ObservationRequest) error {
	// Get the skHash of the given vendor
	skHash, err := ScanDynamoDBWithKeyCondition("vendor", "skHash", "pk", o.PublicKey)

	// Compare requesting skHash with our internal skHash. If they don't match, it's not coming from the vendor
	if *skHash != o.SkHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
		return storage.ErrSkHashDoesNotMatch
	}

	// Initialize AWS session
	svc := utils.SetupAWSSession()

	// Get current time
	currentTime := time.Now()

	// Create an observation to upload to DynamoDB
	d.Observation = Observation{
		Ciphertext:  o.Ciphertext + "|" + o.IV + "|" + o.Attribute,
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

	// Broadcast to all listening clients...
	requestBody, err := json.Marshal(producing.ObservationRequest{
		Attribute:   o.Attribute,
		Ciphertext:  o.Ciphertext + "|" + o.IV + "|" + o.Attribute,
		PublicKey:   o.PublicKey,
		SupertypeID: o.SupertypeID,
		SkHash:      o.SkHash,
		IV:          o.IV,
	})
	if err != nil {
		return storage.ErrMarshaling
	}
	resp, err := http.Post("http://localhost:5001/broadcast", "application/json", bytes.NewBuffer(requestBody))
	if err != nil || resp.StatusCode != 200 {
		log.Printf("error posting: %v\n", err)
	}

	return nil
}
