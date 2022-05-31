package main

import (
	"log"
	"net/http"
	"ws/cmd/internal/handlers"
)

func main() {
	routes := routes()

	log.Println("starting webserver on port 8080")

	log.Println("starting Channel Listener")
	go handlers.ListenToWsChannel()

	_ = http.ListenAndServe(":8080", routes)
}
