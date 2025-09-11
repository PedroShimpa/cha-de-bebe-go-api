package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pedroShimpa/cha-de-bebe-api/models"
	"github.com/pedroShimpa/cha-de-bebe-api/utils"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type CreateEventInput struct {
	Type        models.EventType     `json:"type" binding:"required"`
	Title       string               `json:"title" binding:"required"`
	Description string               `json:"description"`
	PixKey      string               `json:"pix_key"`
	EventDate   time.Time            `json:"event_date" binding:"required"`
	HourStart   time.Time            `json:"hour_start" binding:"required"`
	HourEnd     *time.Time           `json:"hour_end"`
	Address     string               `json:"address" binding:"required"`
	Invited     []CreateInvitedInput `json:"invited"`
	Gifts       []CreateGiftInput    `json:"gifts"`
}

type CreateInvitedInput struct {
	UserID *uint  `json:"user_id,omitempty"`
	Name   string `json:"name" binding:"required"`
}

type CreateGiftInput struct {
	Name            string `json:"name" binding:"required"`
	Link            string `json:"link,omitempty"`
	MaxReservations uint   `json:"max_reservations,omitempty"`
}

type ReserveGiftInput struct {
	UserID      uint `json:"user_id" binding:"required"`
	EventGiftID uint `json:"event_gift_id" binding:"required"`
}

type Controller struct {
	DB *gorm.DB
}

func (ctrl *Controller) CreateEvent(c *gin.Context) {
	var input CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")

	event := models.Event{
		UserID:      userID,
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		PixKey:      input.PixKey,
		EventDate:   input.EventDate,
		HourStart:   input.HourStart,
		HourEnd:     input.HourEnd,
		Address:     input.Address,
	}

	if err := ctrl.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, inv := range input.Invited {
		uuid := utils.GenerateCustomUUID()
		invited := models.EventInvited{
			EventID: event.ID,
			UserID:  inv.UserID,
			Name:    inv.Name,
			UUID:    uuid,
		}
		ctrl.DB.Create(&invited)
	}

	for _, gift := range input.Gifts {
		newGift := models.EventGift{
			EventID:         event.ID,
			Name:            gift.Name,
			Link:            gift.Link,
			MaxReservations: gift.MaxReservations,
		}
		if newGift.MaxReservations == 0 {
			newGift.MaxReservations = 1
		}
		ctrl.DB.Create(&newGift)
	}

	var createdEvent models.Event
	if err := ctrl.DB.Preload("Invited").Preload("Gifts.Reservations").
		First(&createdEvent, event.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": createdEvent})
}

func (ctrl *Controller) GetEvent(c *gin.Context) {
	eventID := c.Param("id")
	var event models.Event
	if err := ctrl.DB.Preload("Invited").Preload("Gifts.Reservations").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"event": event})
}

func (ctrl *Controller) ReserveGift(c *gin.Context) {
	var input ReserveGiftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var gift models.EventGift
	if err := ctrl.DB.Preload("Reservations").First(&gift, input.EventGiftID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Presente não encontrado"})
		return
	}

	if uint(len(gift.Reservations)) >= gift.MaxReservations {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limite de reservas atingido"})
		return
	}

	for _, r := range gift.Reservations {
		if r.UserID == input.UserID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário já reservou este presente"})
			return
		}
	}

	reservation := models.GiftReservation{
		EventGiftID: gift.ID,
		UserID:      input.UserID,
	}
	if err := ctrl.DB.Create(&reservation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível reservar o presente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Presente reservado com sucesso"})
}

func (ctrl *Controller) RespondInvite(c *gin.Context) {
	inviteUUID := c.Param("uuid")
	var invite models.EventInvited

	if err := ctrl.DB.Where("uuid = ?", inviteUUID).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Convite não encontrado"})
		return
	}

	var input struct {
		Accepted bool `json:"accepted" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	invite.Accepted = &input.Accepted
	invite.RespondedAt = &now

	if err := ctrl.DB.Save(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível salvar resposta do convite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invite": invite})
}

func (ctrl *Controller) GetEventByInvite(c *gin.Context) {
	inviteUUID := c.Param("uuid")

	var invite models.EventInvited
	if err := ctrl.DB.Where("uuid = ?", inviteUUID).First(&invite).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Convite não encontrado"})
		return
	}

	var event models.Event
	if err := ctrl.DB.Preload("Gifts.Reservations").
		First(&event, invite.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}

	availableGifts := []models.EventGift{}
	for _, gift := range event.Gifts {
		if len(gift.Reservations) < int(gift.MaxReservations) {
			availableGifts = append(availableGifts, gift)
		}
	}
	event.Gifts = availableGifts
	event.Invited = nil

	c.JSON(http.StatusOK, gin.H{"event": event})
}
