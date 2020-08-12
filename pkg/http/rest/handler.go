package rest

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"

	"github.com/super-type/supertype/pkg/consuming"
	"github.com/super-type/supertype/pkg/dashboard"

	"github.com/gorilla/mux"
	"github.com/super-type/supertype/pkg/authenticating"
	"github.com/super-type/supertype/pkg/producing"
)

func isAuthorized(endpoint func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				// Load environment variable functionality
				err := godotenv.Load()
				if err != nil {
					return nil, authenticating.ErrNotAuthorized
				}

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, authenticating.ErrNotAuthorized
				}

				return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
			})
			if err != nil {
				return
			}

			if token.Valid {
				endpoint(w, r)
			}
		}
	})
}

// Router is the main router for the application
func Router(a authenticating.Service, p producing.Service, c consuming.Service, d dashboard.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/healthcheck", healthcheck()).Methods("GET", "OPTIONS")
	router.HandleFunc("/loginVendor", loginVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/createVendor", createVendor(a)).Methods("POST", "OPTIONS")
	router.HandleFunc("/produce", isAuthorized(produce(p))).Methods("POST", "OPTIONS")
	router.HandleFunc("/consume", isAuthorized(consume(c))).Methods("POST", "OPTIONS")
	router.HandleFunc("/getVendorComparisonMetadata", isAuthorized(getVendorComparisonMetadata(p))).Methods("POST", "OPTIONS")
	router.HandleFunc("/addReencryptionKeys", isAuthorized(addReencryptionKeys(p))).Methods("POST", "OPTIONS")
	router.HandleFunc("/listObservations", isAuthorized(listObservations(d))).Methods("GET", "OPTIONS")

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
		// TODO disable this block when consuming, this is used to enable CORS for local testing
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if (r).Method == "OPTIONS" { // todo we may still want to leave this but unsure
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
