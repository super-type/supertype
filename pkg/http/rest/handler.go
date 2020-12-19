package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/internal/utils"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/dashboard"
	httpUtil "github.com/super-type/supertype/pkg/http"
	"github.com/super-type/supertype/pkg/producing"
)

// Router is the main router for the application
func Router(a authenticating.Service, p producing.Service, c consuming.Service, d dashboard.Service) *mux.Router {
	router := mux.NewRouter()

	// TODO change camel-cased URLs
	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET", "OPTIONS")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/authorized-login-user", authorizedLoginUser(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/loginUser", loginUser(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createUser", createUser(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/consume", consume(c)).Methods("POST", "OPTIONS")
	router.HandleFunc("/produce", produce(p)).Methods("POST", "OPTIONS")
	router.HandleFunc("/listObservations", utils.IsAuthorized(listObservations(d))).Methods("GET", "OPTIONS") // TODO make this list-observations. Do we need isAuthorized()?
	router.HandleFunc("/register-webhook", registerWebhook(d)).Methods("POST", "OPTIONS")                     // TODO do we need isAuthorized()?
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

func authorizedLoginUser(a authenticating.Service) func(w http.ResponseWriter, r *http.Request) {
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

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			return
		}

		result, err := a.AuthorizedLoginUser(user, apiKey)
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

func produce(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
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

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			return
		}

		err = p.Produce(observation, apiKey)
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

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			return
		}

		res, err := c.Consume(observation, apiKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func listObservations(d dashboard.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var observation dashboard.ObservationRequest
		err = decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &observation == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			return
		}

		observations, err := d.ListObservations(observation, apiKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(observations)
	}
}

func registerWebhook(d dashboard.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var webhookRequest dashboard.WebhookRequest
		err = decoder.Decode(&webhookRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &webhookRequest == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			return
		}

		err = d.RegisterWebhook(webhookRequest, apiKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("OK") // todo do something better here
	}
}
