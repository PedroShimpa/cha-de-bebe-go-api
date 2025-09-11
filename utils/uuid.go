package utils

import (
	"math/rand"
	"strings"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateCustomUUID() string {
	rand.Seed(time.Now().UnixNano())
	segments := []int{8, 6, 6, 6}
	sb := strings.Builder{}

	for i, segLen := range segments {
		for j := 0; j < segLen; j++ {
			sb.WriteRune(letters[rand.Intn(len(letters))])
		}
		if i < len(segments)-1 {
			sb.WriteRune('-')
		}
	}

	return sb.String()
}
