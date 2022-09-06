package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// --------------------------------------
// a copy-paste from the ironhook package
// --------------------------------------

func WebhooksHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.URL.Path, "/verification") {
		VerificationHandler(w, r)

	} else if strings.Contains(r.URL.Path, "/notification") {
		NotificationHandler(w, r)

	} else {
		log.Print("Received an unrecognised request: ", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "")
		return
	}
}

func VerificationHandler(w http.ResponseWriter, r *http.Request) {

	qID, ok := r.URL.Query()["id"]

	if !ok || len(qID[0]) < 1 {
		fmt.Fprintf(w, "URL Parameter <id> is missing")
		return
	}

	log.Println("Received a verification request UUID: ", qID[0])
	fmt.Fprint(w, qID[0])
}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {

	// Received a notification
	// -----------------------
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Couldnt make head or tail of this")
		return
	}

	var notif WebhookNotification
	// or
	// var notif ironhook.WebhookNotification
	err = json.Unmarshal(body, &notif)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Couldnt make head or tail of this")
		return
	}

	// Successfully parsed a webhook notification
	// ------------------------------------------
	log.Println("Received a notification UUID: ", notif.EventUUID)
	log.Println(notif.EventUUID, " with topic: ", notif.Topic)
	log.Println(notif.EventUUID, " with body: ", notif.Body)

	fmt.Fprint(w, "Awesome, thanks")
}
