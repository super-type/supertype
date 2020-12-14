package main

import (
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/http/rest"
	"github.com/super-type/supertype/pkg/producing"
	"github.com/super-type/supertype/pkg/storage/dynamo"
)

func main() {
	// Initialize storage
	persistentStorage := new(dynamo.Storage)

	// Initialize services
	authenticator := authenticating.NewService(persistentStorage)
	dashboard := dashboard.NewService(persistentStorage)
	producing := producing.NewService(persistentStorage)
	consuming := consuming.NewService(persistentStorage)

	// Initialize routers and startup server
	httpRouter := rest.Router(authenticator, producing, consuming, dashboard)
	color.Cyan("Starting HTTP server on port 5000...")
	log.Fatal(http.ListenAndServe(":5000", httpRouter))
}
