package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	NomeCompleto  string `json:"nome_completo" gorm:"not null"`
	CPF           string `json:"cpf" gorm:"unique;not null"`
	Email         string `json:"email" gorm:"unique;not null"`
	Whatsapp      string `json:"whatsapp" gorm:"not null"`
	FirabaseToken string `json:"firebase_token" gorm:"null"`
	Senha         string `json:"senha" gorm:"not null"`
}
