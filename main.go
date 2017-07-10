package main

import (
	"github.com/jkobyp/clonereap/api"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", HookServer)
	http.HandleFunc("/api/", api.Handler)
	api.Init()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
