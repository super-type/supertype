package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/super-type/supertype/pkg/authenticating"
)

// IsAuthorized checks the given JWT to ensure vendor is authenticated
func IsAuthorized(endpoint func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
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

// LocalHeaders sets local headers for local running
// todo disable this block when publishing or update headers at least, this is used to enable CORS for local testing
func LocalHeaders(w http.ResponseWriter, r *http.Request) (*json.Decoder, error) {
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	// todo we may still want to leave this but unsure
	if (r).Method == "OPTIONS" {
		return nil, errors.New("OPTIONS")
	}

	decoder := json.NewDecoder(r.Body)
	return decoder, nil
}
