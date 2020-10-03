package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/caching"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/dashboard"
	httpUtil "github.com/super-type/supertype/pkg/http"
	"github.com/super-type/supertype/pkg/producing"
)

// Router is the main router for the application
func Router(a authenticating.Service, p producing.Service, c consuming.Service, d dashboard.Service, cache caching.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET", "OPTIONS")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/loginUser", loginUser(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createUser", createUser(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/produce", produce(p, cache)).Methods("POST", "OPTIONS")
	router.HandleFunc("/getVendorComparisonMetadata", httpUtil.IsAuthorized(getVendorComparisonMetadata(p))).Methods("POST", "OPTIONS")
	router.HandleFunc("/addReencryptionKeys", httpUtil.IsAuthorized(addReencryptionKeys(p))).Methods("POST", "OPTIONS")
	router.HandleFunc("/consume", consume(c)).Methods("POST", "OPTIONS")
	router.HandleFunc("/listObservations", httpUtil.IsAuthorized(listObservations(d))).Methods("GET", "OPTIONS")
	return router
}

func healthcheck() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Healthy.")
	}
}

// loginVendor returns a handler for POST /loginVendor requests
func loginVendor(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err = decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &vendor == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

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
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		// ? should should this be authenticating, or storage?
		var vendor authenticating.Vendor
		err = decoder.Decode(&vendor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &vendor == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

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

func loginUser(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var user authenticating.UserPassword
		err = decoder.Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &user == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		result, err := a.LoginUser(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func createUser(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var user authenticating.UserPassword
		err = decoder.Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &user == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		success, err := a.CreateUser(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(success)
	}
}

func produce(p producing.Service, cache caching.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var observation producing.ObservationRequest
		err = decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &observation == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

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
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var observation consuming.ObservationRequest
		err = decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &observation == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		res, err := c.Consume(observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func addReencryptionKeys(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var rekeyRequest producing.ReencryptionKeysRequest
		err = decoder.Decode(&rekeyRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &rekeyRequest == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

func listObservations(d dashboard.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		observations, err := d.ListObservations()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(observations)
	}
}

func getVendorComparisonMetadata(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var observation producing.ObservationRequest
		err = decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &observation == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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
