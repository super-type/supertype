package redis

import (
	"github.com/fatih/color"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/super-type/supertype/pkg/caching"
	"github.com/super-type/supertype/pkg/storage/dynamo"
)

// Subscribe adds specified attributes to relevant Redis lists
func (d *Storage) Subscribe(c caching.WSObservationRequest) error {
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

// GetSubscribers gets all subscribers of the Redis set
func (d *Storage) GetSubscribers(o caching.ObservationRequest) (*[]string, error) {
	// TODO this should just be EstablishRedisConnection, but we can't access helper function yet
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping().Result()
	if err != nil && pong != "" {
		return nil, err
	}

	val, err := rdb.SMembers(o.Attribute + ":" + o.SupertypeID).Result()
	if err != nil {
		return nil, err
	}

	return &val, nil
}
