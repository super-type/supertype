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
	"github.com/super-type/supertype/pkg/storage/redis"
)

func main() {
	// Set up storage. Can easily add more and interchange as development continues
	d := new(dynamo.Storage)
	r := new(redis.Storage)

	// Initialize services. Can easily add more and interchange as development continues
	authenticator := authenticating.NewService(d)
	producing := producing.NewService(d)
	consumingHTTP := consuming.NewService(d)
	consumingWS := consuming.NewService(r)
	dashboard := dashboard.NewService(d)

	router := rest.Router(authenticator, producing, consumingHTTP, consumingWS, dashboard)
	color.Cyan("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
