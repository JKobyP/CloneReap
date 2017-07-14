package main

import (
	"github.com/jkobyp/clonereap/api"
	"log"
	"net/http"
)

func main() {
	http.Handle("/", reactServer())
	http.Handle("/dist/", assetHandler())
	http.HandleFunc("/api/", api.Handler)
	api.Init()
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func assetHandler() http.Handler {
	return http.FileServer(http.Dir("."))
}
