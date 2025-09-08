package models

import (
	"gorm.io/gorm"
)

type EventGift struct {
	gorm.Model
	EventID         uint              `json:"event_id" gorm:"not null;index"`
	Name            string            `json:"name" gorm:"not null"`
	Link            string            `json:"link,omitempty"`
	MaxReservations uint              `json:"max_reservations" gorm:"default:1"`
	Reservations    []GiftReservation `gorm:"foreignKey:EventGiftID"`
}
