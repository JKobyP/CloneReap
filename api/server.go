package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	cmd := req.URL.Path[len("/api/"):]
	if cmd[0:2] == "pr" {
		PRHandler(w, req)
	} else {
		fmt.Fprintf(w, "Support for %v not implemented!\n", cmd)
		log.Printf("Support for %v not implemented!\n", cmd)
	}
}

func PRHandler(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[len("/api/pr/"):]
	if prid, err := strconv.Atoi(id); err == nil {
		_ = prid
		// TODO figure out what to write back here...
	}
}
