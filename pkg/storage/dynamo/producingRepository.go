package dynamo

import (
	"time"

	"github.com/super-type/supertype/pkg/http/websocket"

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
	skHash, err := GetSkHash(o.PublicKey)

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
		Ciphertext:  o.Ciphertext + "|" + o.IV,
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

	// todo is this kind of crossing bounded contexts? I feel like we shouldn't be calling a handler from a service
	// todo this is where we should get the right re-encyrption data and send it over if we want E2E-encrypted data
	// Broadcast to all listening clients...
	websocket.BroadcastForSpecificPool(o.Attribute+"|"+o.SupertypeID, o.Ciphertext+"|"+o.IV+"|"+o.Attribute)

	return nil
}
