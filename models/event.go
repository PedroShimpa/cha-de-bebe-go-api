package models

import (
	"time"

	"gorm.io/gorm"
)

type EventType string

const (
	Boy        EventType = "boy"
	Girl       EventType = "girl"
	NotDefined EventType = "not_defined"
)

type Event struct {
	gorm.Model
	UserID uint `json:"type" gorm:"not null"`

	Type        EventType `json:"type" gorm:"not null"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	PixKey      string    `json:"pix_key,omitempty"`

	EventDate time.Time  `json:"event_date" gorm:"not null"`
	HourStart time.Time  `json:"hour_start" gorm:"not null"`
	HourEnd   *time.Time `json:"hour_end,omitempty"`
	Address   string     `json:"address" gorm:"not null"`

	Invited []EventInvited `gorm:"foreignKey:EventID"`
	Gifts   []EventGift    `gorm:"foreignKey:EventID"`
}
