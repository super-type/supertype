package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/super-type/supertype/pkg/caching"
)

// Router is the main router for the application
func Router(c caching.Service) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/consumeWS", consumeWS(c)).Methods("GET", "OPTIONS")
	return router
}

func consumeWS(c caching.Service) func(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// todo currently allows any connection no matter the source
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		// todo how do we know which connection is which. Generating our own connection
		//     ID and putting it in Redis doesn't suffice when we have to actually send data to a specific connection....
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cid, err := c.GenerateConnectionID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// todo should we write this as a different message type as well? Like 3 or something?
		err = ws.WriteMessage(1, []byte(*cid))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for {
			// read in a message
			messageType, p, _ := ws.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if messageType == 2 {
				var obs caching.WSObservationRequest
				err := json.Unmarshal(p, &obs)
				if err != nil {
					return
				}
				err = c.Subscribe(obs)
				if err != nil {
					return
				}
				if err := ws.WriteMessage(messageType, []byte("Successfully subscribed")); err != nil {
					return
				}
			}
		}
	}
}
