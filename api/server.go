package api

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	cmd := req.URL.Path[len("/api/"):]
	fmt.Fprintf(w, "Support for %v not implemented!\n", cmd)
}
