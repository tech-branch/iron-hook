package ironhook

import (
	"errors"
	"io"
	"net/url"
	"path"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// append /verification?id=<uuid> to the endpoint
func prepareEndpointForVerification(endpoint, uuid string) (string, error) {

	u, err := prepareEndpointForOperations(endpoint)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "verification")

	q := u.Query()
	q.Set("id", uuid)

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// append /notification to the endpoint
func prepareEndpointForNotification(endpoint string) (string, error) {

	u, err := prepareEndpointForOperations(endpoint)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "notification")

	return u.String(), nil
}

// parse string endpoint into url object
func prepareEndpointForOperations(endpoint string) (*url.URL, error) {

	// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if !u.IsAbs() {
		return nil, ErrEmptyEndpointURLScheme
	}

	scheme_is_not_http := (u.Scheme != "http") && (u.Scheme != "https")
	if scheme_is_not_http {
		return nil, ErrUnsupportedEndpointURLScheme
	}

	if u.Host == "" {
		return nil, ErrEmptyEndpointURL
	}

	return u, nil

}

// verify if the response body holds the expected information
func verifyResponseBodyForVerification(iobody io.Reader, uuid string) error {

	body, err := io.ReadAll(iobody)
	if err != nil {
		return err
	}

	if string(body) != uuid {
		return ErrIncorrectVerificationResponse
		// return fmt.Errorf("incorrect verification response, have %s want %s", string(body), uuid)
	}

	return nil
}

func (s *WebhookEndpointServiceImpl) fetchWebhookEndpointFromDB(endpoint WebhookEndpoint) (*WebhookEndpointDB, error) {

	if endpoint.UUID == uuid.Nil {
		return nil, ErrEmptyEndpointUUID
	}

	var reference_endpoint WebhookEndpointDB

	tx := s.db.First(
		&reference_endpoint,
		"uuid = ?", endpoint.UUID,
	)

	if tx.Error != nil {

		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			s.log.Error(
				"havent found the requested webhook in the database",
				zap.Error(tx.Error),
			)
			return nil, ErrRecordNotFound

		} else {
			s.log.Error(
				"couldnt fetch a webhook from the database",
				zap.Error(tx.Error),
			)
			return nil, tx.Error
		}
	}

	return &reference_endpoint, nil
}

var ErrFailedEndpointVerification error = errors.New(
	"failed to verify an endpoint",
)
var ErrIncorrectEndpointURL error = errors.New(
	"cant accept a URL like this for verification",
)
var ErrEmptyEndpointURL error = errors.New(
	`
	cant accept an empty URL in an Endpoint declaration. 
	Recover by retrying with a non-empty URL in your Endpoint declaration
	`,
)
var ErrEmptyEndpointURLScheme error = errors.New(
	`
	cant accept a URL without http/s.
	you can recover by retrying with "https://"+endpoint
	`,
)
var ErrUnsupportedEndpointURLScheme error = errors.New(
	"cant accept a URL without http/s for verification",
)
var ErrIncorrectVerificationResponse error = errors.New(
	"expected a different endpoint verification response",
)
var ErrRecordNotFound = errors.New(
	`
	Cant find the requested record,
	Recover by retrying with a different identifier
	or, where applicable, by enabling persistence.
	`,
)
var ErrEmptyEndpointUUID error = errors.New(
	`
	cant accept an empty UUID in an Endpoint. 
	Recover by retrying with a non-empty UUID in your Endpoint
	`,
)
