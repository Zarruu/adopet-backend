package handlers

import (
	"net/http"
	"strconv"

	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	Service *services.NotificationService
}

func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{Service: service}
}

func (h *NotificationHandler) HandleGetNotifications(c *gin.Context) {
	userID, _ := c.Get("userID")

	notifications, err := h.Service.GetUserNotifications(userID.(int))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Notifikasi berhasil diambil", notifications)
}

func (h *NotificationHandler) HandleMarkAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID notifikasi tidak valid")
		return
	}

	err = h.Service.MarkAsRead(id, userID.(int))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Notifikasi ditandai sebagai dibaca", nil)
}

func (h *NotificationHandler) HandleGetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")

	count, err := h.Service.GetUnreadCount(userID.(int))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Jumlah notifikasi belum dibaca", gin.H{
		"unread_count": count,
	})
}
