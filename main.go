package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting kwet...")

	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
