package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pedroShimpa/cha-de-bebe-api/models"
	"github.com/pedroShimpa/cha-de-bebe-api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(c *gin.Context, db *gorm.DB) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
	input.Senha = string(hashedPassword)

	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário registrado com sucesso"})
}

func Login(c *gin.Context, db *gorm.DB) {
	var input struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não encontrado"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Senha), []byte(input.Senha)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	token, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func CompletePrfile() {

}
