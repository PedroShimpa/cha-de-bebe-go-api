package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pedroShimpa/cha-de-bebe-api/controllers"
	"github.com/pedroShimpa/cha-de-bebe-api/middlewares"
	"github.com/pedroShimpa/cha-de-bebe-api/models"
	"gorm.io/gorm"
	"time"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// CORS liberando tudo
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	ctrl := controllers.Controller{DB: db}

	// Auth
	r.POST("/register", func(c *gin.Context) { controllers.Register(c, db) })
	r.POST("/login", func(c *gin.Context) { controllers.Login(c, db) })

	// Convites
	r.GET("/invites/:uuid/event", ctrl.GetEventByInvite)
	r.POST("/invites/:uuid/respond", ctrl.RespondInvite)

	// Rotas autenticadas
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		// Eventos
		auth.POST("/events", ctrl.CreateEvent)
		auth.PUT("/events/:id", ctrl.UpdateEvent)
		auth.DELETE("/events/:id", ctrl.DeleteEvent)
		auth.GET("/events/:id", ctrl.GetEvent)

		// Convidados
		auth.POST("/events/:id/invited", ctrl.AddInvited)
		auth.DELETE("/events/:id/invited/:invite_id", ctrl.RemoveInvited)

		// Presentes
		auth.POST("/events/:id/gifts", ctrl.AddGift)
		auth.DELETE("/events/:id/gifts/:gift_id", ctrl.RemoveGift)
		auth.POST("/gifts/reserve", ctrl.ReserveGift)

		// Listar eventos do usuário logado
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
