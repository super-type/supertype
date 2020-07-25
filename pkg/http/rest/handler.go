package rest

import (
	"encoding/json"
	"net/http"

	"github.com/super-type/supertype/pkg/consuming"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/producing"
)

// Router is the main router for the application
func Router(a authenticating.Service, p producing.Service, c consuming.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET", "OPTIONS")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/produce", produce(p)).Methods("POST", "OPTIONS")
	router.HandleFunc("/consume", consume(c)).Methods("POST", "OPTIONS")
	router.HandleFunc("/getVendorComparisonMetadata", getVendorComparisonMetadata(p)).Methods("POST", "OPTIONS")
	router.HandleFunc("/addReencryptionKeys", addReencryptionKeys(p)).Methods("POST", "OPTIONS")

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
		// TODO disable this block when publishing, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
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

		// TODO add null check on values we need. If they're not there, throw a bad request

		result, err := a.LoginVendor(vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func createVendor(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO disable this block when publishing, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
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

		// TODO add null check on values we need. If they're not there, throw a bad request

		keyPair, err := a.CreateVendor(vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		authenticatedVendor, err := a.LoginVendor(vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO do we need a new object for this...?
		result := authenticating.AuthenticatedVendorFirstLogin{
			FirstName:      authenticatedVendor.FirstName,
			LastName:       authenticatedVendor.LastName,
			Email:          authenticatedVendor.Email,
			BusinessName:   authenticatedVendor.BusinessName,
			Username:       authenticatedVendor.Username,
			PublicKey:      keyPair[0],
			PrivateKey:     keyPair[1],
			SupertypeID:    authenticatedVendor.SupertypeID,
			JWT:            authenticatedVendor.JWT,
			AccountBalance: authenticatedVendor.AccountBalance,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func produce(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO disable this block when publishing, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
			return
		}

		decoder := json.NewDecoder(r.Body)

		var observation producing.ObservationRequest
		err := decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO add null check on values we need. If they're not there, throw a bad request

		err = p.Produce(observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode("OK")
	}
}

func consume(c consuming.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO disable this block when consuming, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
			return
		}

		decoder := json.NewDecoder(r.Body)

		var observation consuming.ObservationRequest
		err := decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO add null check on values we need. If they're not there, throw a bad request

		res, err := c.Consume(observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func getVendorComparisonMetadata(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO re-use this code in a util library
		// TODO disable this block when consuming, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
			return
		}

		decoder := json.NewDecoder(r.Body)

		var observation producing.ObservationRequest
		err := decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metadata, err := p.GetVendorComparisonMetadata(observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	}
}

func addReencryptionKeys(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO disable this block when consuming, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
			return
		}

		decoder := json.NewDecoder(r.Body)

		var rekeyRequest producing.ReencryptionKeysRequest
		err := decoder.Decode(&rekeyRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = p.AddReencryptionKeys(rekeyRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("OK")
	}
}
