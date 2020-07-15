package main

import (
	"log"
	"net/http"

	"github.com/super-type/supertype/pkg/authenticating"

	"github.com/fatih/color"
	"github.com/super-type/supertype/pkg/http/rest"
	"github.com/super-type/supertype/pkg/storage/dynamo"
)

func main() {
	// Set up storage. Can easily add more and interchange as development continues
	s := new(dynamo.Storage)

	// Initialize services. Can easily add more and interchange as development continues
	authenticator := authenticating.NewService(s)

	router := rest.Router(authenticator)
	color.Cyan("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
