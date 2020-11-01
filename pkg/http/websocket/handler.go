package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/super-type/supertype/pkg/caching"
	httpUtil "github.com/super-type/supertype/pkg/http"
	"github.com/super-type/supertype/pkg/producing"
)

// Router is the main router for the application
func Router(c caching.Service, p producing.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/consume", consume(c)).Methods("GET", "OPTIONS")
	router.HandleFunc("/broadcast", broadcast(p, c)).Methods("POST", "OPTIONS")
	return router
}

func consume(c caching.Service) func(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// todo currently allows any connection no matter the source
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = conn.WriteMessage(1, []byte("Connected"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Err: %v\n", err)
			return
		}

		var observation caching.WSObservationRequest
		err = json.Unmarshal(p, &observation)
		if err != nil {
			return
		}

		ctx := context.Background()

		var messageJSON *[]byte
		for {
			messageJSON, err = c.Subscribe(ctx, conn, observation.Attribute+"|"+observation.SupertypeID)
			if err != nil {
				return
			}

			err = conn.WriteMessage(2, *messageJSON)
			if err != nil {
				return
			}
		}
	}
}

func broadcast(p producing.Service, c caching.Service) func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.Background()
		c.Publish(ctx, observation.Attribute+"|"+observation.SupertypeID, observation.Ciphertext)

		// todo change this from "OK" to standardized resposne
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("OK")
	}
}
