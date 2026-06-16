package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"adopet-backend/models"
)

type AdoptionService struct {
	DB              *sql.DB
	NotificationSvc *NotificationService
}

func NewAdoptionService(db *sql.DB, notifSvc *NotificationService) *AdoptionService {
	return &AdoptionService{
		DB:              db,
		NotificationSvc: notifSvc,
	}
}

func (s *AdoptionService) SubmitAdoption(adoption *models.Adoption) error {
	// Check if pet exists and is available
	var petStatus string
	var petName string
	err := s.DB.QueryRow("SELECT status, name FROM pets WHERE id = ?", adoption.PetID).Scan(&petStatus, &petName)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("hewan tidak ditemukan")
		}
		return errors.New("gagal memeriksa status hewan")
	}

	if petStatus != "available" {
		return errors.New("hewan ini sudah tidak tersedia untuk diadopsi")
	}

	// Check if user already has a pending adoption for this pet
	var existingCount int
	err = s.DB.QueryRow(
		"SELECT COUNT(*) FROM adoptions WHERE pet_id = ? AND user_id = ? AND status = 'pending'",
		adoption.PetID, adoption.UserID,
	).Scan(&existingCount)
	if err == nil && existingCount > 0 {
		return errors.New("Anda sudah memiliki pengajuan adopsi yang sedang diproses untuk hewan ini")
	}

	result, err := s.DB.Exec(
		"INSERT INTO adoptions (pet_id, user_id, applicant_name, applicant_phone, applicant_email, applicant_address, reason) VALUES (?, ?, ?, ?, ?, ?, ?)",
		adoption.PetID, adoption.UserID, adoption.ApplicantName, adoption.ApplicantPhone,
		adoption.ApplicantEmail, adoption.ApplicantAddress, adoption.Reason,
	)
	if err != nil {
		return errors.New("gagal mengajukan adopsi")
	}

	id, _ := result.LastInsertId()
	adoption.ID = int(id)

	// Update pet status to pending
	_, _ = s.DB.Exec("UPDATE pets SET status = 'pending' WHERE id = ?", adoption.PetID)

	// Create notification for the user
	notif := &models.Notification{
		UserID:  adoption.UserID,
		Title:   "Pengajuan Adopsi Dikirim",
		Message: fmt.Sprintf("Pengajuan adopsi Anda untuk %s telah berhasil dikirim dan sedang menunggu persetujuan.", petName),
		Type:    "info",
		PetName: petName,
	}
	_ = s.NotificationSvc.CreateNotification(notif)

	return nil
}

func (s *AdoptionService) GetUserAdoptions(userID int) ([]models.Adoption, error) {
	rows, err := s.DB.Query(`
		SELECT a.id, a.pet_id, a.user_id, a.applicant_name, a.applicant_phone, a.applicant_email,
			a.applicant_address, a.reason, a.status, a.reviewed_by, a.reviewed_at, a.created_at,
			p.name, p.species, COALESCE(p.image_url, '')
		FROM adoptions a
		LEFT JOIN pets p ON a.pet_id = p.id
		WHERE a.user_id = ?
		ORDER BY a.created_at DESC
	`, userID)
	if err != nil {
		return nil, errors.New("gagal mengambil data adopsi")
	}
	defer rows.Close()

	var adoptions []models.Adoption
	for rows.Next() {
		var a models.Adoption
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime
		var petName, petSpecies, petImageURL sql.NullString

		err := rows.Scan(
			&a.ID, &a.PetID, &a.UserID, &a.ApplicantName, &a.ApplicantPhone, &a.ApplicantEmail,
			&a.ApplicantAddress, &a.Reason, &a.Status, &reviewedBy, &reviewedAt, &a.CreatedAt,
			&petName, &petSpecies, &petImageURL,
		)
		if err != nil {
			continue
		}

		if reviewedBy.Valid {
			v := int(reviewedBy.Int64)
			a.ReviewedBy = &v
		}
		if reviewedAt.Valid {
			a.ReviewedAt = &reviewedAt.Time
		}
		if petName.Valid {
			a.PetName = petName.String
		}
		if petSpecies.Valid {
			a.PetSpecies = petSpecies.String
		}
		if petImageURL.Valid {
			a.PetImageURL = petImageURL.String
		}

		adoptions = append(adoptions, a)
	}

	if adoptions == nil {
		adoptions = []models.Adoption{}
	}

	return adoptions, nil
}

func (s *AdoptionService) GetAllAdoptions(status string) ([]models.Adoption, error) {
	query := `
		SELECT a.id, a.pet_id, a.user_id, a.applicant_name, a.applicant_phone, a.applicant_email,
			a.applicant_address, a.reason, a.status, a.reviewed_by, a.reviewed_at, a.created_at,
			p.name, p.species, COALESCE(p.image_url, '')
		FROM adoptions a
		LEFT JOIN pets p ON a.pet_id = p.id
	`
	args := []interface{}{}

	if status != "" {
		query += " WHERE a.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY a.created_at DESC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, errors.New("gagal mengambil data adopsi")
	}
	defer rows.Close()

	var adoptions []models.Adoption
	for rows.Next() {
		var a models.Adoption
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime
		var petName, petSpecies, petImageURL sql.NullString

		err := rows.Scan(
			&a.ID, &a.PetID, &a.UserID, &a.ApplicantName, &a.ApplicantPhone, &a.ApplicantEmail,
			&a.ApplicantAddress, &a.Reason, &a.Status, &reviewedBy, &reviewedAt, &a.CreatedAt,
			&petName, &petSpecies, &petImageURL,
		)
		if err != nil {
			continue
		}

		if reviewedBy.Valid {
			v := int(reviewedBy.Int64)
			a.ReviewedBy = &v
		}
		if reviewedAt.Valid {
			a.ReviewedAt = &reviewedAt.Time
		}
		if petName.Valid {
			a.PetName = petName.String
		}
		if petSpecies.Valid {
			a.PetSpecies = petSpecies.String
		}
		if petImageURL.Valid {
			a.PetImageURL = petImageURL.String
		}

		adoptions = append(adoptions, a)
	}

	if adoptions == nil {
		adoptions = []models.Adoption{}
	}

	return adoptions, nil
}

func (s *AdoptionService) ApproveAdoption(id, reviewerID int) error {
	var adoption models.Adoption
	var petName string
	err := s.DB.QueryRow(
		`SELECT a.id, a.pet_id, a.user_id, a.status, p.name
		 FROM adoptions a LEFT JOIN pets p ON a.pet_id = p.id WHERE a.id = ?`, id,
	).Scan(&adoption.ID, &adoption.PetID, &adoption.UserID, &adoption.Status, &petName)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pengajuan adopsi tidak ditemukan")
		}
		return errors.New("gagal mengambil data adopsi")
	}

	if adoption.Status != "pending" {
		return errors.New("pengajuan adopsi ini sudah diproses sebelumnya")
	}

	now := time.Now()
	_, err = s.DB.Exec(
		"UPDATE adoptions SET status = 'approved', reviewed_by = ?, reviewed_at = ? WHERE id = ?",
		reviewerID, now, id,
	)
	if err != nil {
		return errors.New("gagal menyetujui adopsi")
	}

	// Update pet status to adopted
	_, _ = s.DB.Exec("UPDATE pets SET status = 'adopted' WHERE id = ?", adoption.PetID)

	// Reject all other pending adoptions for the same pet
	_, _ = s.DB.Exec(
		"UPDATE adoptions SET status = 'rejected', reviewed_by = ?, reviewed_at = ? WHERE pet_id = ? AND id != ? AND status = 'pending'",
		reviewerID, now, adoption.PetID, id,
	)

	// Notify approved user
	notif := &models.Notification{
		UserID:  adoption.UserID,
		Title:   "Adopsi Disetujui! 🎉",
		Message: fmt.Sprintf("Selamat! Pengajuan adopsi Anda untuk %s telah disetujui. Silakan hubungi kami untuk proses selanjutnya.", petName),
		Type:    "approved",
		PetName: petName,
	}
	_ = s.NotificationSvc.CreateNotification(notif)

	// Notify rejected users for the same pet
	rejectedRows, err := s.DB.Query(
		"SELECT user_id FROM adoptions WHERE pet_id = ? AND id != ? AND status = 'rejected' AND reviewed_at = ?",
		adoption.PetID, id, now,
	)
	if err == nil {
		defer rejectedRows.Close()
		for rejectedRows.Next() {
			var rejectedUserID int
			if err := rejectedRows.Scan(&rejectedUserID); err == nil {
				rejectNotif := &models.Notification{
					UserID:  rejectedUserID,
					Title:   "Adopsi Ditolak",
					Message: fmt.Sprintf("Mohon maaf, pengajuan adopsi Anda untuk %s tidak dapat disetujui karena hewan sudah diadopsi oleh pengaju lain.", petName),
					Type:    "rejected",
					PetName: petName,
				}
				_ = s.NotificationSvc.CreateNotification(rejectNotif)
			}
		}
	}

	return nil
}

func (s *AdoptionService) RejectAdoption(id, reviewerID int) error {
	var adoption models.Adoption
	var petName string
	err := s.DB.QueryRow(
		`SELECT a.id, a.pet_id, a.user_id, a.status, p.name
		 FROM adoptions a LEFT JOIN pets p ON a.pet_id = p.id WHERE a.id = ?`, id,
	).Scan(&adoption.ID, &adoption.PetID, &adoption.UserID, &adoption.Status, &petName)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pengajuan adopsi tidak ditemukan")
		}
		return errors.New("gagal mengambil data adopsi")
	}

	if adoption.Status != "pending" {
		return errors.New("pengajuan adopsi ini sudah diproses sebelumnya")
	}

	now := time.Now()
	_, err = s.DB.Exec(
		"UPDATE adoptions SET status = 'rejected', reviewed_by = ?, reviewed_at = ? WHERE id = ?",
		reviewerID, now, id,
	)
	if err != nil {
		return errors.New("gagal menolak adopsi")
	}

	// Check if there are other pending adoptions for this pet
	var pendingCount int
	_ = s.DB.QueryRow("SELECT COUNT(*) FROM adoptions WHERE pet_id = ? AND status = 'pending'", adoption.PetID).Scan(&pendingCount)
	if pendingCount == 0 {
		// No more pending adoptions, set pet back to available
		_, _ = s.DB.Exec("UPDATE pets SET status = 'available' WHERE id = ?", adoption.PetID)
	}

	// Notify user
	notif := &models.Notification{
		UserID:  adoption.UserID,
		Title:   "Adopsi Ditolak",
		Message: fmt.Sprintf("Mohon maaf, pengajuan adopsi Anda untuk %s tidak dapat disetujui. Silakan coba adopsi hewan lainnya.", petName),
		Type:    "rejected",
		PetName: petName,
	}
	_ = s.NotificationSvc.CreateNotification(notif)

	return nil
}
