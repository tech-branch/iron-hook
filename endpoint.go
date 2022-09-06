package ironhook

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type WebhookEndpointStatus int

const (
	Unverified WebhookEndpointStatus = iota
	Suspended
	Verified
	Healthy
)

type WebhookEndpoint struct {
	UUID   uuid.UUID             `json:"uuid"`
	URL    string                `json:"url"`
	Status WebhookEndpointStatus `json:"status"`
}

type WebhookEndpointDB struct {
	gorm.Model
	UUID   uuid.UUID             `gorm:"type:uuid"`
	URL    string                `gorm:"not null"`
	Status WebhookEndpointStatus `gorm:"not null"`
}

func endpointDbToWeb(dbe *WebhookEndpointDB) *WebhookEndpoint {
	return &WebhookEndpoint{
		UUID:   dbe.UUID,
		URL:    dbe.URL,
		Status: dbe.Status,
	}
}

func endpointsDbToWeb(dbes *[]WebhookEndpointDB) *[]WebhookEndpoint {
	web_endpoints := make([]WebhookEndpoint, len(*dbes))
	for i, e := range *dbes {
		web_endpoints[i] = *endpointDbToWeb(&e)
	}
	return &web_endpoints
}
