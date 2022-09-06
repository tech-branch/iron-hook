package main

import (
	"github.com/gofrs/uuid"
	ironhook "github.com/tech-branch/iron-hook"
)

func main() {

	app := newApp()
	aUser := app.newUser(42)

	// aUser registers a webhook endpoint of http://localhost8080
	endpoint_uuid, err := app.RegisterWebhookEndpoint("http://localhost:8080", aUser.ID)
	if err != nil {
		app.log.Fatal(err)
	}

	// verify the provided endpoint.
	// Otherwise, Notify() is not going to work for this endpoint
	_, err = app.webhooks.Verify(ironhook.WebhookEndpoint{
		UUID: endpoint_uuid,
	})
	if err != nil {
		app.log.Fatal(err)
	}

	// some time later...
	// a batch of work has been completed for the user,
	// we want to send a notification about it

	batch_id := uuid.Must(uuid.NewV4()) // a trace

	a_user, err := app.fetchUserByID(42)
	if err != nil {
		app.log.Fatal(err)
	}

	a_users_endpoint := ironhook.WebhookEndpoint{
		UUID: a_user.WebhooksUUID}

	a_users_notification := ironhook.WebhookNotification{
		EventUUID: batch_id,
		Topic:     "Batch completed",
		Body:      "",
	}

	err = app.webhooks.Notify(
		a_users_endpoint,
		a_users_notification,
	)
	if err != nil {
		app.log.Fatal(err)
	}

	// notification sent

}
