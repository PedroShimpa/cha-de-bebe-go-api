package models

import (
	"gorm.io/gorm"
)

type GiftReservation struct {
	gorm.Model
	EventGiftID uint `json:"event_gift_id" gorm:"not null;index"`
	UserID      uint `json:"user_id" gorm:"not null;index"`
}
