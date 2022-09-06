package main

import (
	"log"
	"net/http"
)

// or you can import ironhook.WebhookNotification
type WebhookNotification struct {
	EventUUID string `json:"event_uuid"`
	Topic     string `json:"topic"`
	Body      string `json:"body"`
}

func main() {

	http.HandleFunc("/", WebhooksHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
