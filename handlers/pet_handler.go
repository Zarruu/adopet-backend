package handlers

import (
	"net/http"
	"strconv"

	"adopet-backend/models"
	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type PetHandler struct {
	Service *services.PetService
}

func NewPetHandler(service *services.PetService) *PetHandler {
	return &PetHandler{Service: service}
}

func (h *PetHandler) HandleGetPets(c *gin.Context) {
	search := c.Query("search")
	species := c.Query("species")
	status := c.Query("status")

	pets, err := h.Service.GetAllPets(search, species, status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data hewan berhasil diambil", pets)
}

func (h *PetHandler) HandleGetPet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID hewan tidak valid")
		return
	}

	pet, err := h.Service.GetPetByID(id)
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data hewan berhasil diambil", pet)
}

func (h *PetHandler) HandleCreatePet(c *gin.Context) {
	var req models.CreatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data hewan tidak valid. Pastikan nama, umur, dan ras terisi")
		return
	}

	userID, _ := c.Get("userID")

	pet := &models.Pet{
		Name:         req.Name,
		Age:          req.Age,
		Breed:        req.Breed,
		Species:      req.Species,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		GDriveFileID: req.GDriveFileID,
		Status:       "available",
		CreatedBy:    userID.(int),
	}

	err := h.Service.CreatePet(pet)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Hewan berhasil ditambahkan", pet)
}

func (h *PetHandler) HandleUpdatePet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID hewan tidak valid")
		return
	}

	var req models.UpdatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Data hewan tidak valid")
		return
	}

	pet := &models.Pet{
		ID:           id,
		Name:         req.Name,
		Age:          req.Age,
		Breed:        req.Breed,
		Species:      req.Species,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		GDriveFileID: req.GDriveFileID,
		Status:       req.Status,
	}

	if pet.Status == "" {
		pet.Status = "available"
	}

	err = h.Service.UpdatePet(pet)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Data hewan berhasil diperbarui", pet)
}

func (h *PetHandler) HandleDeletePet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "ID hewan tidak valid")
		return
	}

	err = h.Service.DeletePet(id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Hewan berhasil dihapus", nil)
}
