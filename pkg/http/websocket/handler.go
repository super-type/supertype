package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/super-type/supertype/pkg/caching"
	httpUtil "github.com/super-type/supertype/pkg/http"
	"github.com/super-type/supertype/pkg/producing"
)

// PoolMap is a map of all pools with key of the form <attribute>|<supertypeID>
// todo move this out of memory
var poolMap map[string]httpUtil.Pool

// Router is the main router for the application
func Router(c caching.Service, p producing.Service) *mux.Router {
	router := mux.NewRouter()
	poolMap = make(map[string]httpUtil.Pool)

	// todo maybe we can store PoolIDs in Redis and check to see if it's been started yet - adding to it if so and starting it if not
	color.Cyan("Starting pool...")
	router.HandleFunc("/consume", consume(c)).Methods("GET", "OPTIONS")
	router.HandleFunc("/broadcast", broadcast(p)).Methods("POST", "OPTIONS")
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

		// todo this should probably be moved within caching service once we've moved poolMap out of memory
		if _, ok := poolMap[observation.Attribute+"|"+observation.SupertypeID]; !ok {
			pool := httpUtil.NewPool()
			poolMap[observation.Attribute+"|"+observation.SupertypeID] = *pool
			go pool.Start()

			client := &httpUtil.Client{
				ID:   observation.Attribute + "|" + observation.SupertypeID,
				Conn: conn,
				Pool: pool,
			}

			// We register this client on this specific pool
			pool.Register <- client

			poolMap[observation.Attribute+"|"+observation.SupertypeID] = *pool
			client.Read()
		} else {
			pool := poolMap[observation.Attribute+"|"+observation.SupertypeID]

			client := &httpUtil.Client{
				ID:   observation.Attribute + "|" + observation.SupertypeID,
				Conn: conn,
				Pool: &pool,
			}

			pool.Register <- client
			poolMap[observation.Attribute+"|"+observation.SupertypeID] = pool
			client.Read()
		}
	}
}

func broadcast(p producing.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder, err := httpUtil.LocalHeaders(w, r)
		if err != nil {
			return
		}

		var observation producing.BroadcastRequest
		err = decoder.Decode(&observation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if &observation == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err = p.Broadcast(observation, poolMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// todo change this from "OK" to standardized resposne
		json.NewEncoder(w).Encode("OK")
	}
}
