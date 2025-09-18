package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type InvitePageController struct{}

func (ctrl *InvitePageController) ServePage(c *gin.Context) {
	c.HTML(http.StatusOK, "invite.html", gin.H{})
}
