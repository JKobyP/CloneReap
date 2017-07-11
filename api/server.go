package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	components := strings.Split(req.URL.Path[len("/api/pr/"):], "/")
	if len(components) != 2 {
		log.Printf("invalid url %s", req.URL.Path)
		fmt.Fprintf(w, "invalid url")
		return
	}
	user := components[0]
	project := components[1]
	prs, err := RetrievePrs(user, project)
	if err != nil {
		log.Printf("not found")
		http.NotFound(w, req)
	}
	bytes, err := json.Marshal(prs)
	if err != nil {
		log.Printf("Json error")
		// server error
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", bytes)
}
