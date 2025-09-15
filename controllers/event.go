package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pedroShimpa/cha-de-bebe-api/models"
	"github.com/pedroShimpa/cha-de-bebe-api/utils"
	"gorm.io/gorm"
)

type CreateEventInput struct {
	Type        models.EventType     `json:"type" binding:"required"`
	Title       string               `json:"title" binding:"required"`
	Description string               `json:"description"`
	PixKey      string               `json:"pix_key"`
	EventDate   string               `json:"event_date" binding:"required"`
	HourStart   string               `json:"hour_start" binding:"required"`
	HourEnd     string               `json:"hour_end"`
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
    MaxReservations string `json:"max_reservations,omitempty"` // string
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
		maxRes := uint(1) // padrão
		if gift.MaxReservations != "" {
			if v, err := strconv.Atoi(gift.MaxReservations); err == nil && v > 0 {
				maxRes = uint(v)
			}
		}

		newGift := models.EventGift{
			EventID:         event.ID,
			Name:            gift.Name,
			Link:            gift.Link,
			MaxReservations: maxRes,
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

// ---------------- Editar Evento ----------------
func (ctrl *Controller) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetUint("userID")

	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}

	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas o criador pode editar o evento"})
		return
	}

	var input CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.Title = input.Title
	event.Description = input.Description
	event.PixKey = input.PixKey
	event.EventDate = input.EventDate
	event.HourStart = input.HourStart
	event.HourEnd = input.HourEnd
	event.Address = input.Address
	event.Type = input.Type

	if err := ctrl.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível atualizar o evento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": event})
}

func (ctrl *Controller) DeleteEvent(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetUint("userID")

	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}

	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas o criador pode deletar o evento"})
		return
	}

	if err := ctrl.DB.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar evento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Evento deletado com sucesso"})
}

// ---------------- Convidados ----------------
func (ctrl *Controller) AddInvited(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetUint("userID")

	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}
	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas o criador pode adicionar convidados"})
		return
	}

	var input CreateInvitedInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uuid := utils.GenerateCustomUUID()
	inv := models.EventInvited{
		EventID: event.ID,
		UserID:  input.UserID,
		Name:    input.Name,
		UUID:    uuid,
	}
	if err := ctrl.DB.Create(&inv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar convidado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invited": inv})
}

func (ctrl *Controller) RemoveInvited(c *gin.Context) {
	eventID := c.Param("id")
	inviteID := c.Param("invite_id")
	userID := c.GetUint("userID")

	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}
	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas o criador pode remover convidados"})
		return
	}

	if err := ctrl.DB.Delete(&models.EventInvited{}, inviteID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover convidado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Convidado removido com sucesso"})
}

// ---------------- Presentes ----------------
func (ctrl *Controller) AddGift(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetUint("userID")

	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}
	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas o criador pode adicionar presentes"})
		return
	}

	var input CreateGiftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxRes := uint(1)
	if input.MaxReservations != "" {
		if v, err := strconv.Atoi(input.MaxReservations); err == nil && v > 0 {
			maxRes = uint(v)
		}
	}

	gift := models.EventGift{
		EventID:         event.ID,
		Name:            input.Name,
		Link:            input.Link,
		MaxReservations: maxRes,
	}
	if err := ctrl.DB.Create(&gift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar presente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gift": gift})
}

func (ctrl *Controller) RemoveGift(c *gin.Context) {
	eventID := c.Param("id")
	giftID := c.Param("gift_id")
	userID := c.GetUint("userID")

	var event models.Event
	if err := ctrl.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evento não encontrado"})
		return
	}
	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas o criador pode remover presentes"})
		return
	}

	if err := ctrl.DB.Delete(&models.EventGift{}, giftID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover presente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Presente removido com sucesso"})
}
