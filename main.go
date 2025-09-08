package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pedroShimpa/cha-de-bebe-api/models"
	"github.com/pedroShimpa/cha-de-bebe-api/routes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Nenhum .env encontrado, usando variáveis do sistema")
	}

	dbDSN := os.Getenv("DB_DSN")
	port := os.Getenv("PORT")

	db, err := gorm.Open(mysql.Open(dbDSN), &gorm.Config{})
	if err != nil {
		panic("falha ao conectar ao banco de dados")
	}

	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.UserProfile{})
	db.AutoMigrate(&models.Event{})
	db.AutoMigrate(&models.EventGift{})
	db.AutoMigrate(&models.EventInvited{})
	db.AutoMigrate(&models.GiftReservation{})

	r := gin.Default()
	routes.SetupRoutes(r, db)

	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
