package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pedroShimpa/cha-de-bebe-api/controllers"
	"github.com/pedroShimpa/cha-de-bebe-api/middlewares"
	"github.com/pedroShimpa/cha-de-bebe-api/models"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	ctrl := controllers.Controller{DB: db}

	r.POST("/register", func(c *gin.Context) { controllers.Register(c, db) })
	r.POST("/login", func(c *gin.Context) { controllers.Login(c, db) })

	r.POST("/invites/:uuid/event", ctrl.GetEventByInvite)
	r.POST("/invites/:uuid/respond", ctrl.RespondInvite)

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())

	{
		auth.POST("/events", ctrl.CreateEvent)
		auth.GET("/events/:id", ctrl.GetEvent)
		auth.POST("/gifts/reserve", ctrl.ReserveGift)

		auth.GET("/events", func(c *gin.Context) {
			userID := c.GetUint("userID")
			var events []models.Event

			if err := db.Preload("Invited").Preload("Gifts.Reservations").
				Where("user_id = ?", userID).
				Find(&events).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"events": events})
		})

	}
}
