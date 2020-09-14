package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatih/color"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/storage"
)

// Consume returns all observations at the requested attribute for the specified Supertype entity
func (d *Storage) Consume(c consuming.ObservationRequest) (*[]consuming.ObservationResponse, error) {
	skHash, err := GetSkHash(c.PublicKey)
	if err != nil {
		return nil, err
	}
	// Compare requesting skHash with our internal skHash. If they don't match, it's not coming from the vendor
	if *skHash != c.SkHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
		return nil, storage.ErrSkHashDoesNotMatch
	}

	// Initialize AWS session
	svc := utils.SetupAWSSession()

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
	}

	return &observations, nil
}

// Subscribe does nothing here
func (d *Storage) Subscribe(c consuming.WSObservationRequest) error {
	return nil
}

// GenerateConnectionID does nothing here
func (d *Storage) GenerateConnectionID() (*string, error) {
	return nil, nil
}
