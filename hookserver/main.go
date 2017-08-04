package hookserver

import (
	"github.com/jkobyp/clonereap/api"
	"log"
	"net/http"
)

func Main(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HookServer)
	mux.HandleFunc("/api/", api.Handler)
	api.Init()
	log.Fatal(http.ListenAndServe(port, mux))
}
