package main

import (
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/caching"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/http/rest"
	"github.com/super-type/supertype/pkg/http/websocket"
	"github.com/super-type/supertype/pkg/producing"
	"github.com/super-type/supertype/pkg/storage/dynamo"
	"github.com/super-type/supertype/pkg/storage/redis"
)

func main() {
	// Initialize storage
	persistentStorage := new(dynamo.Storage)
	cacheStorage := new(redis.Storage)

	// Initialize services
	authenticator := authenticating.NewService(persistentStorage)
	dashboard := dashboard.NewService(persistentStorage)
	producing := producing.NewService(persistentStorage)
	consuming := consuming.NewService(persistentStorage)
	cache := caching.NewService(cacheStorage)

	// Initialize routers and startup server
	httpRouter := rest.Router(authenticator, producing, consuming, dashboard, cache)
	wsRouter := websocket.Router(cache, producing)
	// todo is this the best way to do this?
	// source: https://gist.github.com/filewalkwithme/24363472e7424bbe7028
	go func() {
		color.Cyan("Starting HTTP server on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", httpRouter))
	}()
	color.Cyan("Starting WebSocket server on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", wsRouter))
}
