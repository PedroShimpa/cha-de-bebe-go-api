package utils

import (
	"math/rand"
	"strings"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// gera um slug curto, tipo "cha-abc123"
func GenerateSlug() string {
	rand.Seed(time.Now().UnixNano())
	sb := strings.Builder{}
	for i := 0; i < 6; i++ { // slug de 6 chars randômicos
		sb.WriteRune(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}
