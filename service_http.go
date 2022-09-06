package ironhook

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Aggregates the http webhook functionality into a test Server
// for testing and mocks.
func mockHttpWebhooksServerForTests() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(webhooksHandler))
}

// A primitive router for testing and mocks.
func webhooksHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.URL.Path, "/verification") {
		VerificationHandler(w, r)

	} else if strings.Contains(r.URL.Path, "/notification") {
		mockNotificationHandler(w, r)

	} else {
		log.Print("Received an unrecognised request: ", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "")
		return
	}
}

// Handles incoming Verification requests.
// To be used from the perspective of the webhook receiver.
func VerificationHandler(w http.ResponseWriter, r *http.Request) {

	q_id, ok := r.URL.Query()["id"]

	if !ok || len(q_id[0]) < 1 {
		fmt.Fprintf(w, "Url Param 'id' is missing")
		return
	}

	r_uuid := q_id[0]
	fmt.Fprint(w, r_uuid)
}

// Notification handler for testing and mocks.
func mockNotificationHandler(w http.ResponseWriter, r *http.Request) {

	// Received a notification
	// -----------------------
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Didnt understand your, sorry")
		return
	}

	var notif WebhookNotification
	err = json.Unmarshal(body, &notif)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Didnt understand your, sorry")
		return
	}

	// Successfully parsed a webhook notification
	// ------------------------------------------
	log.Println("Received a notification with UUID: ", notif.EventUUID)
	log.Println(notif.EventUUID, " with topic: ", notif.Topic)
	log.Println(notif.EventUUID, " with body: ", notif.Body)

	fmt.Fprint(w, "Awesome, thanks")
}
