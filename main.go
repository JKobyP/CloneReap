package main

import (
	"github.com/jkobyp/clonereap/api"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", reactServer)
	http.HandleFunc("/api/", api.Handler)
	api.Init()
	log.Fatal(http.ListenAndServe(":8000", nil))
}
