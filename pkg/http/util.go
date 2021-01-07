package http

import (
	"encoding/json"
	"errors"
	"net/http"
)

// LocalHeaders sets local headers for local running
// todo disable this block when publishing or update headers at least, this is used to enable CORS for local testing
func LocalHeaders(w http.ResponseWriter, r *http.Request) (*json.Decoder, error) {
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key, Token")
	// todo we may still want to leave this but unsure
	if (r).Method == "OPTIONS" {
		return nil, errors.New("OPTIONS")
	}

	decoder := json.NewDecoder(r.Body)
	return decoder, nil
}
