package services

import (
	"database/sql"
	"errors"

	"adopet-backend/models"
)

type NotificationService struct {
	DB *sql.DB
}

func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{DB: db}
}

func (s *NotificationService) GetUserNotifications(userID int) ([]models.Notification, error) {
	rows, err := s.DB.Query(
		"SELECT id, user_id, title, message, type, COALESCE(pet_name, ''), is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, errors.New("gagal mengambil notifikasi")
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.PetName, &n.IsRead, &n.CreatedAt)
		if err != nil {
			continue
		}
		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []models.Notification{}
	}

	return notifications, nil
}

func (s *NotificationService) MarkAsRead(id, userID int) error {
	result, err := s.DB.Exec(
		"UPDATE notifications SET is_read = TRUE WHERE id = ? AND user_id = ?",
		id, userID,
	)
	if err != nil {
		return errors.New("gagal menandai notifikasi sebagai dibaca")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("notifikasi tidak ditemukan")
	}

	return nil
}

func (s *NotificationService) GetUnreadCount(userID int) (int, error) {
	var count int
	err := s.DB.QueryRow(
		"SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = FALSE",
		userID,
	).Scan(&count)
	if err != nil {
		return 0, errors.New("gagal mengambil jumlah notifikasi belum dibaca")
	}
	return count, nil
}

func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	result, err := s.DB.Exec(
		"INSERT INTO notifications (user_id, title, message, type, pet_name) VALUES (?, ?, ?, ?, ?)",
		notification.UserID, notification.Title, notification.Message, notification.Type, notification.PetName,
	)
	if err != nil {
		return errors.New("gagal membuat notifikasi")
	}

	id, _ := result.LastInsertId()
	notification.ID = int(id)

	return nil
}
