package models

import "time"

type Pet struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Age          string    `json:"age"`
	Breed        string    `json:"breed"`
	Species      string    `json:"species"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	GDriveFileID string    `json:"gdrive_file_id,omitempty"`
	Status       string    `json:"status"`
	CreatedBy    int       `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreatePetRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Age         string `json:"age" binding:"required"`
	Breed       string `json:"breed" binding:"required"`
	Species     string `json:"species"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	GDriveFileID string `json:"gdrive_file_id"`
}

type UpdatePetRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Age         string `json:"age" binding:"required"`
	Breed       string `json:"breed" binding:"required"`
	Species     string `json:"species"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	GDriveFileID string `json:"gdrive_file_id"`
	Status      string `json:"status"`
}
