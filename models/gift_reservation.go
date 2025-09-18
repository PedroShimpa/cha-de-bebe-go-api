package models

import (
	"gorm.io/gorm"
)

type GiftReservation struct {
	gorm.Model
	EventGiftID uint   `json:"event_gift_id" gorm:"not null;index"`
	InviteUUID  string `json:"invite_uuid" gorm:"not null;index"`
}
