package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	GDriveService *services.GDriveService
}

func NewUploadHandler(gdriveSvc *services.GDriveService) *UploadHandler {
	return &UploadHandler{GDriveService: gdriveSvc}
}

func (h *UploadHandler) HandleUploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "File gambar diperlukan. Gunakan field 'image'")
		return
	}
	defer file.Close()

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		utils.Error(c, http.StatusBadRequest, "Ukuran file maksimal 5MB")
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		utils.Error(c, http.StatusBadRequest, "Format file tidak didukung. Gunakan JPG, PNG, GIF, atau WEBP")
		return
	}

	fileID, webURL, err := h.GDriveService.UploadFile(file, header.Filename)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengunggah gambar: "+err.Error())
		return
	}

	responseData := gin.H{
		"file_id":   fileID,
		"image_url": webURL,
	}

	utils.Success(c, http.StatusOK, "Gambar berhasil diunggah", responseData)
}
