package handlers

import (
	"net/http"
	"strconv"

	"adopet-backend/models"
	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type AdoptionHandler struct {
	Service *services.AdoptionService
}

func NewAdoptionHandler(service *services.AdoptionService) *AdoptionHandler {
	return &AdoptionHandler{Service: service}
}

func (h *AdoptionHandler) HandleSubmitAdoption(c *gin.Context) {
	var req models.SubmitAdoptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data adopsi tidak valid. Pastikan semua field terisi dengan benar")
		return
	}

	userID, _ := c.Get("userID")

	adoption := &models.Adoption{
		PetID:            req.PetID,
		UserID:           userID.(int),
		ApplicantName:    req.ApplicantName,
		ApplicantPhone:   req.ApplicantPhone,
		ApplicantEmail:   req.ApplicantEmail,
		ApplicantAddress: req.ApplicantAddress,
		Reason:           req.Reason,
		Status:           "pending",
	}

	err := h.Service.SubmitAdoption(adoption)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Pengajuan adopsi berhasil dikirim", adoption)
}

func (h *AdoptionHandler) HandleGetMyAdoptions(c *gin.Context) {
	userID, _ := c.Get("userID")

	adoptions, err := h.Service.GetUserAdoptions(userID.(int))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data adopsi berhasil diambil", adoptions)
}

func (h *AdoptionHandler) HandleGetAllAdoptions(c *gin.Context) {
	status := c.Query("status")

	adoptions, err := h.Service.GetAllAdoptions(status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data adopsi berhasil diambil", adoptions)
}

func (h *AdoptionHandler) HandleApproveAdoption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID adopsi tidak valid")
		return
	}

	reviewerID, _ := c.Get("userID")

	err = h.Service.ApproveAdoption(id, reviewerID.(int))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Adopsi berhasil disetujui", nil)
}

func (h *AdoptionHandler) HandleRejectAdoption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID adopsi tidak valid")
		return
	}

	reviewerID, _ := c.Get("userID")

	err = h.Service.RejectAdoption(id, reviewerID.(int))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Adopsi berhasil ditolak", nil)
}
