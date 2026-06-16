package handlers

import (
	"net/http"
	"strconv"

	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

func (h *UserHandler) HandleGetUsers(c *gin.Context) {
	role := c.Query("role")

	users, err := h.Service.GetAllUsers(role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data pengguna berhasil diambil", users)
}

func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required,min=2,max=100"`
		Role     string `json:"role" binding:"required,oneof=user editor admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data pengguna tidak valid. Pastikan semua field terisi dengan benar")
		return
	}

	user, err := h.Service.CreateUser(req.Username, req.Email, req.Password, req.Name, req.Role)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Pengguna berhasil dibuat", user)
}

func (h *UserHandler) HandleUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID pengguna tidak valid")
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required,min=2,max=100"`
		Email    string `json:"email" binding:"required,email"`
		Role     string `json:"role" binding:"required,oneof=user editor admin"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data pengguna tidak valid")
		return
	}

	err = h.Service.UpdateUser(id, req.Name, req.Email, req.Role, req.IsActive)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Pengguna berhasil diperbarui", nil)
}

func (h *UserHandler) HandleDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID pengguna tidak valid")
		return
	}

	// Prevent self-deletion
	currentUserID, _ := c.Get("userID")
	if currentUserID.(int) == id {
		utils.Error(c, http.StatusBadRequest, "Tidak dapat menghapus akun sendiri")
		return
	}

	err = h.Service.DeleteUser(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Pengguna berhasil dihapus", nil)
}
