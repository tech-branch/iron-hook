package ironhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebhookEndpointService interface {
	Create(WebhookEndpoint) (WebhookEndpoint, error)
	UpdateURL(WebhookEndpoint) (WebhookEndpoint, error)
	Verify(WebhookEndpoint) (WebhookEndpoint, error)
	Get(WebhookEndpoint) (WebhookEndpoint, error)
	Delete(WebhookEndpoint) error
	Notify(WebhookEndpoint, WebhookNotification) error
	LastNotificationSent(WebhookEndpoint) (WebhookNotification, error)
	ListEndpoints() (*[]WebhookEndpoint, error)
}

type WebhookEndpointServiceImpl struct {
	WebhookEndpointService
	db   *gorm.DB
	log  *zap.Logger
	http *http.Client
}

// Creates a new Webhook service, connects to a database and applies migrations
func NewWebhookService(custom_http_client *http.Client) (WebhookEndpointService, error) {

	// Viper Config
	// ------------
	initConfig()

	// Logging
	// -------
	logger, err := newLoggerAtLevel(
		viper.GetString("log_level"),
	)
	if err != nil {
		return nil, err
	}

	// DB instance
	// -----------
	logger.Info("Getting a SQL store")
	db, err := databaseConnection(
		viper.GetString("db_engine"),
		viper.GetString("db_dsn"),
	)
	if err != nil {
		return nil, err
	}

	// DB migrations
	// -------------
	logger.Info("Applying database migrations")
	err = db.AutoMigrate(
		&WebhookEndpointDB{},
		&WebhookNotificationDB{},
	)
	if err != nil {
		return nil, err
	}

	// Universal HTTP client
	// ---------------------
	var http_client *http.Client
	if custom_http_client == nil {
		http_client = &http.Client{
			Timeout: time.Second * 3,
		}
	} else {
		http_client = custom_http_client
	}

	// Pulling it all together
	// -----------------------
	logger.Info("Pulling together a new Webhooks service")
	return &WebhookEndpointServiceImpl{
		db:   db,
		log:  logger,
		http: http_client,
	}, nil
}

// Creates a new unverified Webhook Endpoint. The next step would be to run
// the verification process on this endpoint.
//
// verified_endpoint, err := service.Verify(endpoint)
//
// Otherwise you will not be able to send Notifiations to the endpoint.
func (s *WebhookEndpointServiceImpl) Create(endpoint WebhookEndpoint) (WebhookEndpoint, error) {

	endpoint.Status = Unverified

	if endpoint.UUID == uuid.Nil {
		endpoint.UUID = uuid.Must(uuid.NewV4())
	}

	if endpoint.URL == "" {
		// recover by retrying with a non-empty endpoint URL
		return endpoint, ErrEmptyEndpointURL
	}

	db_endpoint := WebhookEndpointDB{
		UUID:   endpoint.UUID,
		URL:    endpoint.URL,
		Status: endpoint.Status,
	}

	s.log.Info(
		"Creating a new endpoint",
		zap.String("URL", db_endpoint.URL),
		zap.String("UUID", db_endpoint.UUID.String()),
	)
	defer func() {
		s.log.Info("A new endpoint created",
			zap.String("URL", db_endpoint.URL),
			zap.String("UUID", db_endpoint.UUID.String()))
	}()

	tx := s.db.Create(&db_endpoint)
	return endpoint, tx.Error
}

// Updates the endpoint with a new URL. The endpoint is then switched to Unverified
// so you will have to verify it again.
//
// verified_endpoint, err := service.Verify(endpoint)
//
// Otherwise you will not be able to send Notifiations to the endpoint.
//
// endpoint.UUID is used to find the webhook in the database.
func (s *WebhookEndpointServiceImpl) UpdateURL(endpoint WebhookEndpoint) (WebhookEndpoint, error) {

	model_endpoint, err := s.fetchWebhookEndpointFromDB(endpoint)
	if err != nil {
		return endpoint, err
	}
	if model_endpoint == nil {
		// that shouldn't happen if there are no error
		return endpoint, ErrInternalProcessingError
	}

	if endpoint.URL == model_endpoint.URL {
		s.log.Info("URL is same as before, nothing to do", zap.String("UUID", endpoint.UUID.String()))
		return endpoint, nil
	}

	s.log.Info("updating the URL and Status", zap.String("UUID", endpoint.UUID.String()))

	model_endpoint.URL = endpoint.URL
	model_endpoint.Status = Unverified

	tx := s.db.Save(&model_endpoint)
	return endpoint, tx.Error
}

// Verifies user's control over the provided endpoint.
// Runs a simple check to see if the endpoint responds
// with an expected answer.
//
// given a request like below
//
// <endpoint.URL>/verification?id=abcd
//
// The verification process expects to see "abcd" in the response body.
func (s *WebhookEndpointServiceImpl) Verify(endpoint WebhookEndpoint) (WebhookEndpoint, error) {
	model_endpoint, err := s.fetchWebhookEndpointFromDB(endpoint)
	if err != nil {
		return endpoint, err
	}
	if model_endpoint == nil {
		// this shouldnt happen if there are no errors
		return endpoint, ErrInternalProcessingError
	}

	if model_endpoint.Status > Suspended {
		// Verified = 2, Healthy = 3
		s.log.Info("endpoint already verified")
		return endpoint, nil
	}

	url, err := prepareEndpointForVerification(model_endpoint.URL, model_endpoint.UUID.String())
	if err != nil {
		return endpoint, err
	}

	var resp *http.Response
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	resp, err = client.Get(url)
	if err != nil {
		return endpoint, err
	}
	defer resp.Body.Close()

	verificationResult := verifyResponseBodyForVerification(
		resp.Body,
		endpoint.UUID.String(),
	)
	if verificationResult != nil {
		s.log.Warn(
			"Failed endpoint verification",
			zap.String("UUID", endpoint.UUID.String()),
			zap.Error(verificationResult))
		return endpoint, ErrFailedEndpointVerification
	}
	s.log.Info("successfully verified an endpoint",
		zap.String("UUID", endpoint.UUID.String()))

	// mark the URL unverified
	model_endpoint.Status = Verified
	endpoint.Status = Verified
	tx := s.db.Save(&model_endpoint)
	return endpoint, tx.Error
}

// Deletes the indicated Endpoint
//
// endpoint.UUID is used to find the webhook in the database.
func (s *WebhookEndpointServiceImpl) Delete(endpoint WebhookEndpoint) error {
	model_endpoint, err := s.fetchWebhookEndpointFromDB(endpoint)
	if err != nil {
		return err
	}
	if model_endpoint == nil {
		// shouldnt happen if there are no errors
		return ErrInternalProcessingError
	}

	tx := s.db.Delete(&model_endpoint)
	return tx.Error
}

// Fetches the indicated Endpoint from the database
//
// endpoint.UUID is used to find the webhook in the database.
func (s *WebhookEndpointServiceImpl) Get(endpoint WebhookEndpoint) (WebhookEndpoint, error) {

	model_endpoint, err := s.fetchWebhookEndpointFromDB(endpoint)
	if err != nil {
		return endpoint, err
	}
	if model_endpoint == nil {
		// shouldnt happen if there are no errors
		return endpoint, ErrInternalProcessingError
	}

	simplified_webhook_endpoint := WebhookEndpoint{
		UUID:   model_endpoint.UUID,
		URL:    model_endpoint.URL,
		Status: model_endpoint.Status,
	}

	return simplified_webhook_endpoint, nil
}

// Notify sends a Notification to a verified Endpoint.
//
// Notification's Topic and Body can be empty
func (s *WebhookEndpointServiceImpl) Notify(endpoint WebhookEndpoint, notification WebhookNotification) error {

	ref_endpoint, err := s.Get(endpoint)
	if err != nil {
		return err
	}

	if ref_endpoint.Status < Verified {
		// recover by running service.Verify(endpoint) first
		return ErrEndpointNotYetActivated
	}

	jsonval, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	final_url, err := prepareEndpointForNotification(ref_endpoint.URL)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("GET", final_url, bytes.NewBuffer(jsonval))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := s.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// TODO: introduce toggle for notifications persistence
	// save the notification
	db_notifiaction := WebhookNotificationDB{
		EventUUID:    notification.EventUUID,
		Topic:        notification.Topic,
		Body:         notification.Body,
		EndpointUUID: ref_endpoint.UUID,
	}
	tx := s.db.Create(&db_notifiaction)
	if tx.Error != nil {
		return tx.Error
	}

	// TODO: perhaps worth narrowing down
	if response.StatusCode >= 400 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return ErrFailedNotifyingTheEndpoint
		}
		s.log.Warn(
			"notifiaction returned HTTP error code:",
			zap.Int("StatusCode", response.StatusCode),
			zap.String("Body", string(body)),
		)
		return ErrFailedNotifyingTheEndpoint
	}
	return nil
}

func (s *WebhookEndpointServiceImpl) LastNotificationSent(endpoint WebhookEndpoint) (WebhookNotification, error) {

	if endpoint.UUID == uuid.Nil {
		return WebhookNotification{}, ErrEmptyEndpointUUID
	}

	var db_notifiaction WebhookNotificationDB
	tx := s.db.Last(&db_notifiaction, "endpoint_uuid = ?", endpoint.UUID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			s.log.Error("couldnt find any notifications in the database for provided endpoint", zap.Error(tx.Error))
			return WebhookNotification{}, ErrRecordNotFound
		} else {
			s.log.Error("couldnt fetch a notification from the database", zap.Error(tx.Error))
			return WebhookNotification{}, tx.Error
		}
	}

	return WebhookNotification{
		EventUUID: db_notifiaction.EventUUID,
		Topic:     db_notifiaction.Topic,
		Body:      db_notifiaction.Body,
	}, nil

}

func (s *WebhookEndpointServiceImpl) ListEndpoints() (*[]WebhookEndpoint, error) {

	var db_endpoints []WebhookEndpointDB

	tx := s.db.Find(&db_endpoints)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return endpointsDbToWeb(&db_endpoints), nil
}

var ErrEndpointNotYetActivated error = errors.New(
	`the endpoint youre trying to notify hasnt yet been activated.
	recover by running service.Verify(endpoint) first`,
)
var ErrFailedNotifyingTheEndpoint error = errors.New(
	`the endpoint youre trying to notify returned an error.
	This is perhaps an external issue, recovering might require 
	checking if the request body is parsed correctly and on both ends`,
)
var ErrInternalProcessingError error = errors.New(
	`an internal error occured that shouldn't have happened.
	Please submit an issue with as much detail as possible`,
)
