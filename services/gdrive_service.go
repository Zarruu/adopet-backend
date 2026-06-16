package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GDriveService struct {
	service  *drive.Service
	folderID string
	isLocal  bool
}

var gdriveInstance *GDriveService

func NewGDriveService() *GDriveService {
	if gdriveInstance != nil {
		return gdriveInstance
	}
	gdriveInstance = &GDriveService{}
	return gdriveInstance
}

func (g *GDriveService) Init() error {
	credJSON := os.Getenv("GDRIVE_CREDENTIALS_JSON")
	folderID := os.Getenv("GDRIVE_FOLDER_ID")

	if credJSON == "" {
		log.Println("GDRIVE_CREDENTIALS_JSON tidak diset, menggunakan penyimpanan lokal")
		g.isLocal = true

		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return fmt.Errorf("gagal membuat direktori uploads: %w", err)
		}

		return nil
	}

	// Validate JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal([]byte(credJSON), &jsonCheck); err != nil {
		return fmt.Errorf("GDRIVE_CREDENTIALS_JSON bukan JSON yang valid: %w", err)
	}

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsJSON([]byte(credJSON)))
	if err != nil {
		return fmt.Errorf("gagal membuat Google Drive service: %w", err)
	}

	g.service = srv
	g.folderID = folderID
	g.isLocal = false

	log.Println("Google Drive service berhasil diinisialisasi")
	return nil
}

func (g *GDriveService) IsLocal() bool {
	return g.isLocal
}

func (g *GDriveService) UploadFile(file multipart.File, filename string) (string, string, error) {
	if g.isLocal {
		return g.uploadLocal(file, filename)
	}
	return g.uploadToDrive(file, filename)
}

func (g *GDriveService) uploadLocal(file multipart.File, filename string) (string, string, error) {
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("gagal membuat direktori uploads: %w", err)
	}

	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, newFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("gagal membuat file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", "", fmt.Errorf("gagal menyimpan file: %w", err)
	}

	// Return file path relative to server
	webURL := "/uploads/" + newFilename

	return newFilename, webURL, nil
}

func (g *GDriveService) uploadToDrive(file multipart.File, filename string) (string, string, error) {
	if g.service == nil {
		return "", "", fmt.Errorf("Google Drive service belum diinisialisasi")
	}

	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	mimeType := getMimeType(ext)

	driveFile := &drive.File{
		Name:     newFilename,
		MimeType: mimeType,
	}

	if g.folderID != "" {
		driveFile.Parents = []string{g.folderID}
	}

	res, err := g.service.Files.Create(driveFile).
		Media(file).
		Fields("id, webViewLink, webContentLink").
		Do()
	if err != nil {
		return "", "", fmt.Errorf("gagal mengunggah file ke Google Drive: %w", err)
	}

	// Make file publicly accessible
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = g.service.Permissions.Create(res.Id, permission).Do()
	if err != nil {
		log.Printf("Peringatan: gagal mengatur permission file %s: %v", res.Id, err)
	}

	webURL := fmt.Sprintf("https://drive.google.com/uc?export=view&id=%s", res.Id)

	return res.Id, webURL, nil
}

func (g *GDriveService) DeleteFile(fileID string) error {
	if g.isLocal {
		return g.deleteLocal(fileID)
	}
	return g.deleteFromDrive(fileID)
}

func (g *GDriveService) deleteLocal(filename string) error {
	filePath := filepath.Join("uploads", filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("gagal menghapus file lokal: %w", err)
	}
	return nil
}

func (g *GDriveService) deleteFromDrive(fileID string) error {
	if g.service == nil {
		return fmt.Errorf("Google Drive service belum diinisialisasi")
	}

	if fileID == "" {
		return nil
	}

	err := g.service.Files.Delete(fileID).Do()
	if err != nil {
		return fmt.Errorf("gagal menghapus file dari Google Drive: %w", err)
	}

	return nil
}

func getMimeType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
