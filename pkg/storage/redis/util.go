package redis

import "github.com/go-redis/redis"

// EstablishRedisConnection establishes a basic Redis connection
func EstablishRedisConnection() (*redis.Client, error) {
	// Example new client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping().Result()
	if err != nil && pong != "" {
		return nil, err
	}
	return rdb, nil
}
