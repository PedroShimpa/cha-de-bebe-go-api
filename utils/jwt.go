package utils

import (
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var JwtSecret []byte

func init() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "segredo_super_secreto" // fallback
    }
    JwtSecret = []byte(secret)
}

func GenerateToken(userID uint) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 72).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(JwtSecret)
}
