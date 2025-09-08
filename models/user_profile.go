package models

import (
	"gorm.io/gorm"
)

type UserProfile struct {
	gorm.Model
	UserId      int  `json:"user_id" gorm:"not null"`
	IsOrganizer bool `json:"is_organizer" gorm:"unique;not null; default 0"`
}
