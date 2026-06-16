package services

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"adopet-backend/models"
	"adopet-backend/utils"
)

type AuthService struct {
	DB *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) Register(username, email, password, name string) (*models.User, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return nil, errors.New("gagal memeriksa username")
	}
	if count > 0 {
		return nil, errors.New("username sudah digunakan")
	}

	err = s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
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

	result, err := s.DB.Exec(
		"INSERT INTO users (username, email, password_hash, name, role) VALUES (?, ?, ?, ?, 'user')",
		username, email, hashedPassword, name,
	)
	if err != nil {
		return nil, errors.New("gagal mendaftarkan pengguna")
	}

	id, _ := result.LastInsertId()

	user := &models.User{
		ID:       int(id),
		Username: username,
		Email:    email,
		Name:     name,
		Role:     "user",
		IsActive: true,
	}

	return user, nil
}

func (s *AuthService) Login(email, password, ipAddress, userAgent string) (*models.User, string, error) {
	user := &models.User{}
	var photoURL sql.NullString
	var lastLoginAt sql.NullTime

	err := s.DB.QueryRow(
		"SELECT id, username, email, password_hash, name, role, photo_url, is_active, last_login_at, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &photoURL, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", errors.New("email atau password salah")
		}
		return nil, "", errors.New("gagal mencari pengguna")
	}

	if !user.IsActive {
		return nil, "", errors.New("akun Anda telah dinonaktifkan. Hubungi administrator")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, "", errors.New("email atau password salah")
	}

	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, "", errors.New("gagal membuat token autentikasi")
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = s.DB.Exec(
		"INSERT INTO sessions (user_id, token_hash, ip_address, user_agent, expires_at) VALUES (?, ?, ?, ?, ?)",
		user.ID, tokenHash, ipAddress, userAgent, expiresAt,
	)
	if err != nil {
		return nil, "", errors.New("gagal membuat sesi")
	}

	_, _ = s.DB.Exec("UPDATE users SET last_login_at = ? WHERE id = ?", time.Now(), user.ID)

	user.PasswordHash = ""

	return user, token, nil
}

func (s *AuthService) Logout(userID int, token string) error {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	_, err := s.DB.Exec("DELETE FROM sessions WHERE user_id = ? AND token_hash = ?", userID, tokenHash)
	if err != nil {
		return errors.New("gagal menghapus sesi")
	}
	return nil
}

func (s *AuthService) GetProfile(userID int) (*models.User, error) {
	user := &models.User{}
	var photoURL sql.NullString
	var lastLoginAt sql.NullTime

	err := s.DB.QueryRow(
		"SELECT id, username, email, name, role, photo_url, is_active, last_login_at, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Name, &user.Role, &photoURL, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("pengguna tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil profil pengguna")
	}

	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

func (s *AuthService) UpdateProfile(userID int, name, photoURL string) error {
	_, err := s.DB.Exec(
		"UPDATE users SET name = ?, photo_url = ?, updated_at = ? WHERE id = ?",
		name, photoURL, time.Now(), userID,
	)
	if err != nil {
		return errors.New("gagal memperbarui profil")
	}
	return nil
}

func (s *AuthService) ChangePassword(userID int, oldPassword, newPassword string) error {
	var currentHash string
	err := s.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&currentHash)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	if !utils.CheckPassword(oldPassword, currentHash) {
		return errors.New("password lama tidak sesuai")
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("gagal mengenkripsi password baru")
	}

	_, err = s.DB.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", newHash, time.Now(), userID)
	if err != nil {
		return errors.New("gagal memperbarui password")
	}

	return nil
}

func (s *AuthService) CreateDefaultAdmin(username, email, password, name string) error {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return fmt.Errorf("gagal memeriksa admin: %w", err)
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi password admin: %w", err)
	}

	_, err = s.DB.Exec(
		"INSERT INTO users (username, email, password_hash, name, role) VALUES (?, ?, ?, ?, 'admin')",
		username, email, hashedPassword, name,
	)
	if err != nil {
		return fmt.Errorf("gagal membuat admin default: %w", err)
	}

	return nil
}
