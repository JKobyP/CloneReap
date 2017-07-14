package main

import (
	"log"
	"net/http"
)

func reactServer() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
		log.Printf("Serving index.html")
	}
	return http.HandlerFunc(fn)
}
