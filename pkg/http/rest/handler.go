package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/super-type/supertype/internal/utils"

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

		// TODO this should probably be a separate goroutine?
		// TODO this part should probably all be in the websocket handler...
		obs := caching.ObservationRequest{
			Attribute:   observation.Attribute,
			SupertypeID: observation.SupertypeID,
			PublicKey:   observation.PublicKey,
			SkHash:      observation.SkHash,
		}

		// Get all connections subscribed to the given attribute
		subscribers, err := cache.GetSubscribers(obs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("subscribers: %v\n", subscribers)

		// todo this maybe shouldn't be a string array but a map with the pk as key and object containing the metadata as the value
		// todo this is really pivotal around encryption. If we don't have encryption, this becomes a lot faster
		// todo if we keep encryption, we'll need to load all re-encryption info into local memory on startup
		// todo without encryption, we can just do this all within the Produce function using the incoming data
		var pkList []string

		// Iterate through each connection
		for _, subscriber := range *subscribers {
			// Break pipe-delimited subscriber into string array
			subscriberMetadata := strings.Split(subscriber, "|")

			// Get the connection ID
			// connSubscriber := subscriberMetadata[0]

			// Get the public key.
			pkSubscriber := subscriberMetadata[1]

			if !utils.Contains(pkList, pkSubscriber) {
				// Get the necessary connection metadata from DynamoDB
			} else {
				// Get the necessary connection metadata from pkList
			}

			// Attach that metadata to an outgoing response via WebSocket

			// Send that data via WebSocket to the appropriate connection
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
