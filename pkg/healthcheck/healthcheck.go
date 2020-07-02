package healthcheck

import (
	"fmt"
	"net/http"
)

// Healthcheck simply prints whether or not the API is hit
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "app healthy...\n")
}
