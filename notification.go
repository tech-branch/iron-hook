package ironhook

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type WebhookNotification struct {
	EventUUID uuid.UUID `json:"event_uuid"`
	Topic     string    `json:"topic"`
	Body      string    `json:"body"`
}

type WebhookNotificationDB struct {
	gorm.Model
	EventUUID    uuid.UUID `gorm:"type:uuid"`
	Topic        string    `gorm:"not null"`
	Body         string    `gorm:"not null"`
	EndpointUUID uuid.UUID `gorm:"type:uuid"`
}
