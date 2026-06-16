package handlers

import (
	"net/http"

	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	UserService *services.UserService
}

func NewDashboardHandler(userService *services.UserService) *DashboardHandler {
	return &DashboardHandler{UserService: userService}
}

func (h *DashboardHandler) HandleGetDashboard(c *gin.Context) {
	stats, err := h.UserService.GetDashboardStats()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil statistik dashboard")
		return
	}

	utils.Success(c, http.StatusOK, "Statistik dashboard berhasil diambil", stats)
}

func (h *DashboardHandler) HandleGetActiveUsers(c *gin.Context) {
	activeUsers, err := h.UserService.GetActiveUsers()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data pengguna aktif berhasil diambil", activeUsers)
}
