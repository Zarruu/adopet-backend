package handlers

import (
	"net/http"

	"adopet-backend/models"
	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data registrasi tidak valid. Pastikan semua field terisi dengan benar")
		return
	}

	user, err := h.Service.Register(req.Username, req.Email, req.Password, req.Name)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Registrasi berhasil", user)
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Email dan password diperlukan")
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	user, token, err := h.Service.Login(req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Login berhasil", gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) HandleLogout(c *gin.Context) {
	userID, _ := c.Get("userID")
	token, _ := c.Get("token")

	err := h.Service.Logout(userID.(int), token.(string))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Logout berhasil", nil)
}

func (h *AuthHandler) HandleGetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := h.Service.GetProfile(userID.(int))
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Profil berhasil diambil", user)
}

func (h *AuthHandler) HandleUpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data profil tidak valid")
		return
	}

	err := h.Service.UpdateProfile(userID.(int), req.Name, req.PhotoURL)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Profil berhasil diperbarui", nil)
}

func (h *AuthHandler) HandleChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data password tidak valid. Password baru minimal 6 karakter")
		return
	}

	err := h.Service.ChangePassword(userID.(int), req.OldPassword, req.NewPassword)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Password berhasil diubah", nil)
}
