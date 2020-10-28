package dynamo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	httpUtil "github.com/super-type/supertype/pkg/http"
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
	requestBody, err := json.Marshal(producing.BroadcastRequest{
		Attribute:   o.Attribute,
		Ciphertext:  o.Ciphertext + "|" + o.IV + "|" + o.Attribute,
		PublicKey:   o.PublicKey,
		SupertypeID: o.SupertypeID,
		SkHash:      o.SkHash,
		IV:          o.IV,
		PoolID:      o.Attribute + "|" + o.SupertypeID,
	})
	if err != nil {
		return storage.ErrMarshaling
	}
	resp, err := http.Post("http://localhost:8081/broadcast", "application/json", bytes.NewBuffer(requestBody))
	if err != nil || resp.StatusCode != 200 {
		log.Printf("error posting: %v\n", err)
	}

	return nil
}

// Broadcast sends a message to all members of a specific pool
func (d *Storage) Broadcast(b producing.BroadcastRequest, poolMap map[string]httpUtil.Pool) error {
	pool := poolMap[b.PoolID]
	for client := range pool.Clients {
		message := httpUtil.Message{
			Type: 2,
			Body: b.Ciphertext,
		}

		messageJSON, err := json.Marshal(message)
		if err != nil {
			return err
		}

		err = client.Conn.WriteMessage(2, messageJSON)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
