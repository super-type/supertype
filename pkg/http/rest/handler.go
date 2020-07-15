package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/authenticating"
)

// Router is the main router for the application
func Router(a authenticating.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET", "OPTIONS")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST", "OPTIONS")

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
		w.Header().Set("Content-Type", "application/json")

		// TODO disable this block when publishing, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" {
			return
		}

		decoder := json.NewDecoder(r.Body)

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err := decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := a.LoginVendor(vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO log user in after creating account
		// TODO create re-encryption keys between this user and others

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result) // TODO return the JWT from login here...
	}
}
