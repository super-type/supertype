package router

import (
	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/healthcheck"
)

// Router is the main router for the application
func Router() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck.Healthcheck).Methods("GET")

	return router
}
