package models

import (
	"time"

	"gorm.io/gorm"
)

type EventInvited struct {
	gorm.Model
	EventID     uint       `json:"event_id" gorm:"not null;index"`
	UserID      *uint      `json:"user_id,omitempty"`
	Name        string     `json:"name" gorm:"not null"`
	Accepted    *bool      `json:"accepted,omitempty"`
	UUID        string     `json:"uuid" gorm:"unique;not null"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
