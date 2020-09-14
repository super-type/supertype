package redis

import (
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/storage/dynamo"
)

// User is a user to add to Redis
type User struct {
	User   string
	Vendor string
	CID    string
}

// Consume does nothing here
func (d *Storage) Consume(c consuming.ObservationRequest) (*[]consuming.ObservationResponse, error) {
	return nil, nil
}

// GenerateConnectionID generates connection IDs for WebSocket connections
func (d *Storage) GenerateConnectionID() (*string, error) {
	// Example new client
	rdb, err := EstablishRedisConnection()
	if err != nil {
		return nil, err
	}
	color.Cyan("Connected to Redis cache for GenerateConnectionId...")

	cid := uuid.New().String()
	for ok := true; ok; ok = rdb.Exists(cid).Val() > 0 {
		// todo this always triggers once... is there a way to avoid this?
		color.Yellow("Connection ID already exists... retrying...")
		cid = uuid.New().String()
	}

	return &cid, nil
}

// Subscribe adds specified attributes to relevant Redis lists
func (d *Storage) Subscribe(c consuming.WSObservationRequest) error {
	skHash, err := dynamo.GetSkHash(c.PublicKey)
	if err != nil || skHash == nil {
		return err
	}

	// Compare requesting skHash with our internal skHash. If they don't match, it's not coming from the vendor
	if *skHash != c.SkHash {
		color.Red("!!! Vendor secret key hashes do no match - potential malicious attempt !!!")
		return err
	}

	// Example new client
	rdb, err := EstablishRedisConnection()
	if err != nil {
		return err
	}
	color.Cyan("Connected to Redis cache for Subscribe...")

	rdb.SAdd(c.Attribute+":"+c.SupertypeID, c.Cid+"|"+c.PublicKey).Err()
	if err != nil {
		return err
	}

	return nil
}
