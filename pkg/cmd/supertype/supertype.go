package supertype

import (
	"fmt"
	"log"
	"net/http"

	"github.com/super-type/supertype/pkg/router"
)

// RunApplication starts to the application
func RunApplication() {
	router := router.Router()
	fmt.Println("starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
