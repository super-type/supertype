package main

import (
	"log"
	"net/http"

	"github.com/super-type/supertype/pkg/dashboard"
	"github.com/super-type/supertype/pkg/storage/dynamo"

	"github.com/super-type/supertype/pkg/consuming"

	"github.com/super-type/supertype/pkg/producing"

	"github.com/super-type/supertype/pkg/authenticating"

	"github.com/fatih/color"
	"github.com/super-type/supertype/pkg/http/rest"
)

func main() {
	// Set up storage. Can easily add more and interchange as development continues
	s := new(dynamo.Storage)

	// Initialize services. Can easily add more and interchange as development continues
	authenticator := authenticating.NewService(s)
	producing := producing.NewService(s)
	consuming := consuming.NewService(s)
	dashboard := dashboard.NewService(s)

	router := rest.Router(authenticator, producing, consuming, dashboard)
	color.Cyan("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
