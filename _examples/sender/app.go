package main

import (
	"errors"
	"log"
	"sync"

	"github.com/gofrs/uuid"
	ironhook "github.com/tech-branch/iron-hook"
)

// -----------------------------------------
// this is just a primitive application stub
// it serves to represent some basic actions
// -----------------------------------------

// a simplified representation of an app
type App struct {
	log      log.Logger
	users    sync.Map
	webhooks ironhook.WebhookEndpointService
}

func newApp() App {

	webhook_svc, err := ironhook.NewWebhookService(nil)
	if err != nil {
		log.Fatal(err)
	}

	return App{
		log:      *log.Default(),
		users:    sync.Map{},
		webhooks: webhook_svc,
	}

}

// a simplified representation of a user
type User struct {
	ID           int
	WebhooksUUID uuid.UUID
}

func (app *App) RegisterWebhookEndpoint(url string, user_id int) (uuid.UUID, error) {

	endpoint, err := app.webhooks.Create(
		ironhook.WebhookEndpoint{
			URL: url,
		})
	if err != nil {
		return endpoint.UUID, err
	}

	// load user
	usr, err := app.fetchUserByID(user_id)
	if err != nil {
		return endpoint.UUID, err
	}

	// change userdata
	usr.WebhooksUUID = endpoint.UUID
	// save
	app.users.Store(usr.ID, usr)

	return endpoint.UUID, nil

}

func (app *App) fetchUserByID(id int) (User, error) {

	raw_user, ok := app.users.Load(id)
	if !ok {
		return User{}, ErrCouldntLoadUser
	}
	a_user, ok := raw_user.(User)
	if !ok {
		return User{}, ErrCouldntParseUser
	}
	return a_user, nil

}

func (app *App) newUser(id int) User {

	u := User{
		ID: id,
	}

	app.users.Store(u.ID, u)
	return u

}

var ErrCouldntLoadUser = errors.New(
	`
	Couldn't load user from the database,
	You might want to try with another ID.
	`,
)

var ErrCouldntParseUser = errors.New(
	`
	Couldn't parse the user data fetched from the database,
	You might want to try with another ID.
	`,
)
