package services

import (
	"database/sql"
	"errors"
	"time"

	"adopet-backend/models"
	"adopet-backend/utils"
)

type UserService struct {
	AuthDB *sql.DB
	PetsDB *sql.DB
}

func NewUserService(authDB, petsDB *sql.DB) *UserService {
	return &UserService{
		AuthDB: authDB,
		PetsDB: petsDB,
	}
}

func (s *UserService) GetAllUsers(role string) ([]models.User, error) {
	query := "SELECT id, username, email, name, role, COALESCE(photo_url, ''), is_active, last_login_at, created_at, updated_at FROM users"
	args := []interface{}{}

	if role != "" {
		query += " WHERE role = ?"
		args = append(args, role)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.AuthDB.Query(query, args...)
	if err != nil {
		return nil, errors.New("gagal mengambil data pengguna")
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Name, &user.Role,
			&user.PhotoURL, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	return users, nil
}

func (s *UserService) CreateUser(username, email, password, name, role string) (*models.User, error) {
	var count int
	err := s.AuthDB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return nil, errors.New("gagal memeriksa username")
	}
	if count > 0 {
		return nil, errors.New("username sudah digunakan")
	}

	err = s.AuthDB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return nil, errors.New("gagal memeriksa email")
	}
	if count > 0 {
		return nil, errors.New("email sudah terdaftar")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	if role == "" {
		role = "user"
	}

	result, err := s.AuthDB.Exec(
		"INSERT INTO users (username, email, password_hash, name, role) VALUES (?, ?, ?, ?, ?)",
		username, email, hashedPassword, name, role,
	)
	if err != nil {
		return nil, errors.New("gagal membuat pengguna")
	}

	id, _ := result.LastInsertId()

	user := &models.User{
		ID:       int(id),
		Username: username,
		Email:    email,
		Name:     name,
		Role:     role,
		IsActive: true,
	}

	return user, nil
}

func (s *UserService) UpdateUser(id int, name, email, role string, isActive bool) error {
	var count int
	err := s.AuthDB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&count)
	if err != nil || count == 0 {
		return errors.New("pengguna tidak ditemukan")
	}

	// Check email uniqueness
	err = s.AuthDB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", email, id).Scan(&count)
	if err != nil {
		return errors.New("gagal memeriksa email")
	}
	if count > 0 {
		return errors.New("email sudah digunakan oleh pengguna lain")
	}

	_, err = s.AuthDB.Exec(
		"UPDATE users SET name = ?, email = ?, role = ?, is_active = ?, updated_at = ? WHERE id = ?",
		name, email, role, isActive, time.Now(), id,
	)
	if err != nil {
		return errors.New("gagal memperbarui pengguna")
	}

	return nil
}

func (s *UserService) DeleteUser(id int) error {
	var role string
	err := s.AuthDB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pengguna tidak ditemukan")
		}
		return errors.New("gagal memeriksa pengguna")
	}

	if role == "admin" {
		var adminCount int
		_ = s.AuthDB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
		if adminCount <= 1 {
			return errors.New("tidak dapat menghapus admin terakhir")
		}
	}

	_, err = s.AuthDB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return errors.New("gagal menghapus pengguna")
	}

	return nil
}

func (s *UserService) GetActiveUsers() ([]models.ActiveUserInfo, error) {
	rows, err := s.AuthDB.Query(`
		SELECT u.id, u.username, u.name, u.email, u.role, COALESCE(s.ip_address, ''), s.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.expires_at > NOW()
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, errors.New("gagal mengambil data pengguna aktif")
	}
	defer rows.Close()

	var activeUsers []models.ActiveUserInfo
	for rows.Next() {
		var info models.ActiveUserInfo
		err := rows.Scan(
			&info.UserID, &info.Username, &info.Name, &info.Email,
			&info.Role, &info.IPAddress, &info.LoginAt, &info.ExpiresAt,
		)
		if err != nil {
			continue
		}
		activeUsers = append(activeUsers, info)
	}

	if activeUsers == nil {
		activeUsers = []models.ActiveUserInfo{}
	}

	return activeUsers, nil
}

func (s *UserService) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Total users
	_ = s.AuthDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Active sessions
	_ = s.AuthDB.QueryRow("SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()").Scan(&stats.ActiveSessions)

	// Total pets
	_ = s.PetsDB.QueryRow("SELECT COUNT(*) FROM pets").Scan(&stats.TotalPets)

	// Available pets
	_ = s.PetsDB.QueryRow("SELECT COUNT(*) FROM pets WHERE status = 'available'").Scan(&stats.AvailablePets)

	// Adopted pets
	_ = s.PetsDB.QueryRow("SELECT COUNT(*) FROM pets WHERE status = 'adopted'").Scan(&stats.AdoptedPets)

	// Total adoptions
	_ = s.PetsDB.QueryRow("SELECT COUNT(*) FROM adoptions").Scan(&stats.TotalAdoptions)

	// Pending adoptions
	_ = s.PetsDB.QueryRow("SELECT COUNT(*) FROM adoptions WHERE status = 'pending'").Scan(&stats.PendingAdoptions)

	return stats, nil
}
