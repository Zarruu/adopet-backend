package services

import (
	"database/sql"
	"errors"
	"strings"

	"adopet-backend/models"
)

type PetService struct {
	DB *sql.DB
}

func NewPetService(db *sql.DB) *PetService {
	return &PetService{DB: db}
}

func (s *PetService) GetAllPets(search, species, status string) ([]models.Pet, error) {
	query := "SELECT id, name, age, breed, species, COALESCE(description,''), COALESCE(image_url,''), COALESCE(gdrive_file_id,''), status, created_by, created_at, updated_at FROM pets WHERE 1=1"
	args := []interface{}{}

	if search != "" {
		query += " AND (name LIKE ? OR breed LIKE ? OR description LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if species != "" {
		query += " AND species = ?"
		args = append(args, species)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, errors.New("gagal mengambil data hewan")
	}
	defer rows.Close()

	var pets []models.Pet
	for rows.Next() {
		var pet models.Pet
		err := rows.Scan(
			&pet.ID, &pet.Name, &pet.Age, &pet.Breed, &pet.Species,
			&pet.Description, &pet.ImageURL, &pet.GDriveFileID,
			&pet.Status, &pet.CreatedBy, &pet.CreatedAt, &pet.UpdatedAt,
		)
		if err != nil {
			continue
		}
		pets = append(pets, pet)
	}

	if pets == nil {
		pets = []models.Pet{}
	}

	return pets, nil
}

func (s *PetService) GetPetByID(id int) (*models.Pet, error) {
	pet := &models.Pet{}

	err := s.DB.QueryRow(
		"SELECT id, name, age, breed, species, COALESCE(description,''), COALESCE(image_url,''), COALESCE(gdrive_file_id,''), status, created_by, created_at, updated_at FROM pets WHERE id = ?",
		id,
	).Scan(
		&pet.ID, &pet.Name, &pet.Age, &pet.Breed, &pet.Species,
		&pet.Description, &pet.ImageURL, &pet.GDriveFileID,
		&pet.Status, &pet.CreatedBy, &pet.CreatedAt, &pet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("hewan tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data hewan")
	}

	return pet, nil
}

func (s *PetService) CreatePet(pet *models.Pet) error {
	if strings.TrimSpace(pet.Species) == "" {
		pet.Species = "Anjing"
	}
	if strings.TrimSpace(pet.Status) == "" {
		pet.Status = "available"
	}

	result, err := s.DB.Exec(
		"INSERT INTO pets (name, age, breed, species, description, image_url, gdrive_file_id, status, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		pet.Name, pet.Age, pet.Breed, pet.Species, pet.Description,
		pet.ImageURL, pet.GDriveFileID, pet.Status, pet.CreatedBy,
	)
	if err != nil {
		return errors.New("gagal menambahkan hewan")
	}

	id, _ := result.LastInsertId()
	pet.ID = int(id)

	return nil
}

func (s *PetService) UpdatePet(pet *models.Pet) error {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM pets WHERE id = ?", pet.ID).Scan(&count)
	if err != nil || count == 0 {
		return errors.New("hewan tidak ditemukan")
	}

	if strings.TrimSpace(pet.Species) == "" {
		pet.Species = "Anjing"
	}

	_, err = s.DB.Exec(
		"UPDATE pets SET name = ?, age = ?, breed = ?, species = ?, description = ?, image_url = ?, gdrive_file_id = ?, status = ? WHERE id = ?",
		pet.Name, pet.Age, pet.Breed, pet.Species, pet.Description,
		pet.ImageURL, pet.GDriveFileID, pet.Status, pet.ID,
	)
	if err != nil {
		return errors.New("gagal memperbarui data hewan")
	}

	return nil
}

func (s *PetService) DeletePet(id int) error {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM pets WHERE id = ?", id).Scan(&count)
	if err != nil || count == 0 {
		return errors.New("hewan tidak ditemukan")
	}

	_, err = s.DB.Exec("DELETE FROM pets WHERE id = ?", id)
	if err != nil {
		return errors.New("gagal menghapus hewan")
	}

	return nil
}

func (s *PetService) GetPetGDriveFileID(petID int) (string, error) {
	var fileID sql.NullString
	err := s.DB.QueryRow("SELECT gdrive_file_id FROM pets WHERE id = ?", petID).Scan(&fileID)
	if err != nil {
		return "", err
	}
	if fileID.Valid {
		return fileID.String, nil
	}
	return "", nil
}
