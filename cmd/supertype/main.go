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
	color.Cyan("Starting HTTP server on port 5000...")
	log.Fatal(http.ListenAndServe(":5000", httpRouter))
}
