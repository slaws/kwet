package main

import (
	"fmt"
	"net/http"

	nats "github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
)

var nc nats.Conn

func main() {
	log.Info("Starting kwet...")

	router := NewRouter()
	nc, err := NatsConnect("test-cluster", "1")
	if err != nil {
		log.Fatal(err)
	}

	nc.QueueSubscribe("foo", "coin", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	}, nats.DeliverAllAvailable(), nats.DurableName("testprog"))

	log.Fatal(http.ListenAndServe(":8080", router))
}
