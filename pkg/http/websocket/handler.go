package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/super-type/supertype/pkg/caching"
	httpUtil "github.com/super-type/supertype/pkg/http"
)

// PoolMap is a map of all pools with key of the form <attribute>|<supertypeID>
// todo move this out of memory
var poolMap map[string]httpUtil.Pool

// BroadcastForSpecificPool sends a message to all members of a specific pool
// todo move this into util once poolMap is out of memory
func BroadcastForSpecificPool(poolID string, data string) {
	pool := poolMap[poolID]
	for client := range pool.Clients {
		message := httpUtil.Message{
			Type: 2,
			Body: data,
		}

		err := client.Conn.WriteJSON(message)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// Router is the main router for the application
func Router(c caching.Service) *mux.Router {
	router := mux.NewRouter()
	poolMap = make(map[string]httpUtil.Pool)

	// todo maybe we can store PoolIDs in Redis and check to see if it's been started yet - adding to it if so and starting it if not
	color.Cyan("Starting pool...")
	router.HandleFunc("/consume", consume(c)).Methods("GET", "OPTIONS")
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

		// todo should we write this as a different message type as well? Like 3 or something?
		err = conn.WriteMessage(1, []byte("Subscribed"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// client.Read()
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
		} else {
			pool := poolMap[observation.Attribute+"|"+observation.SupertypeID]

			client := &httpUtil.Client{
				ID:   observation.Attribute + "|" + observation.SupertypeID,
				Conn: conn,
				Pool: &pool,
			}

			pool.Register <- client
			poolMap[observation.Attribute+"|"+observation.SupertypeID] = pool
		}
	}
}
