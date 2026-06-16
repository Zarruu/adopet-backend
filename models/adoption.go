package models

import "time"

type Adoption struct {
	ID               int        `json:"id"`
	PetID            int        `json:"pet_id"`
	UserID           int        `json:"user_id"`
	ApplicantName    string     `json:"applicant_name"`
	ApplicantPhone   string     `json:"applicant_phone"`
	ApplicantEmail   string     `json:"applicant_email"`
	ApplicantAddress string     `json:"applicant_address"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	ReviewedBy       *int       `json:"reviewed_by"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	CreatedAt        time.Time  `json:"created_at"`

	// Joined fields
	PetName     string `json:"pet_name,omitempty"`
	PetSpecies  string `json:"pet_species,omitempty"`
	PetImageURL string `json:"pet_image_url,omitempty"`
}

type SubmitAdoptionRequest struct {
	PetID            int    `json:"pet_id" binding:"required"`
	ApplicantName    string `json:"applicant_name" binding:"required,min=2,max=100"`
	ApplicantPhone   string `json:"applicant_phone" binding:"required"`
	ApplicantEmail   string `json:"applicant_email" binding:"required,email"`
	ApplicantAddress string `json:"applicant_address" binding:"required"`
	Reason           string `json:"reason" binding:"required"`
}
