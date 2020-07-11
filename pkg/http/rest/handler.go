package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/authenticating"
)

// Router is the main router for the application
func Router(a authenticating.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST")

	return router
}

func healthcheck() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("healthy...")
	}
}

// loginVendor returns a handler for POST /loginVendor requests
func loginVendor(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err := decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		jwt, err := a.LoginVendor(vendor)
		if err != nil {
			fmt.Printf("Error logging vendor in\n")
			// TODO return here with error message
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwt)
	}
}

func createVendor(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err := decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := a.CreateVendor(vendor)
		if err != nil {
			fmt.Printf("Error creating vendor\n")
			// TODO return here with error message
		}

		// TODO log user in after creating account
		// TODO create re-encryption keys between this user and others

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result) // TODO return the JWT from login here...
	}
}
