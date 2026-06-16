package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	AuthDB *sql.DB
	PetsDB *sql.DB
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak ada file .env ditemukan, menggunakan environment variables sistem")
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func ConnectDatabases() (*sql.DB, *sql.DB, error) {
	authDB, err := connectAndMigrate(
		GetEnv("DB_AUTH_HOST", "localhost"),
		GetEnv("DB_AUTH_PORT", "3306"),
		GetEnv("DB_AUTH_USER", "root"),
		GetEnv("DB_AUTH_PASSWORD", ""),
		GetEnv("DB_AUTH_NAME", "adopet_auth"),
		migrateAuthDB,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal koneksi ke database auth: %w", err)
	}

	petsDB, err := connectAndMigrate(
		GetEnv("DB_PETS_HOST", "localhost"),
		GetEnv("DB_PETS_PORT", "3306"),
		GetEnv("DB_PETS_USER", "root"),
		GetEnv("DB_PETS_PASSWORD", ""),
		GetEnv("DB_PETS_NAME", "adopet_pets"),
		migratePetsDB,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal koneksi ke database pets: %w", err)
	}

	AuthDB = authDB
	PetsDB = petsDB

	return authDB, petsDB, nil
}

func connectAndMigrate(host, port, user, password, dbName string, migrateFn func(*sql.DB) error) (*sql.DB, error) {
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&multiStatements=true", user, password, host, port)
	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi root: %w", err)
	}

	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		rootDB.Close()
		return nil, fmt.Errorf("gagal membuat database %s: %w", dbName, err)
	}
	rootDB.Close()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi ke %s: %w", dbName, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping ke %s: %w", dbName, err)
	}

	log.Printf("Berhasil terhubung ke database: %s", dbName)

	if err := migrateFn(db); err != nil {
		return nil, fmt.Errorf("gagal migrasi %s: %w", dbName, err)
	}

	return db, nil
}

func migrateAuthDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		name VARCHAR(100) NOT NULL,
		role ENUM('user', 'editor', 'admin') DEFAULT 'user',
		photo_url TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		last_login_at TIMESTAMP NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_users_email (email),
		INDEX idx_users_username (username),
		INDEX idx_users_role (role)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		token_hash VARCHAR(255) NOT NULL,
		ip_address VARCHAR(45),
		user_agent TEXT,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		INDEX idx_sessions_user_id (user_id),
		INDEX idx_sessions_token_hash (token_hash),
		INDEX idx_sessions_expires_at (expires_at)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel auth: %w", err)
	}

	log.Println("Migrasi database auth berhasil")
	return nil
}

func migratePetsDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS pets (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		age VARCHAR(50) NOT NULL,
		breed VARCHAR(100) NOT NULL,
		species VARCHAR(50) DEFAULT 'Anjing',
		description TEXT,
		image_url TEXT,
		gdrive_file_id VARCHAR(255),
		status ENUM('available', 'adopted', 'pending') DEFAULT 'available',
		created_by INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_pets_status (status),
		INDEX idx_pets_species (species),
		INDEX idx_pets_created_by (created_by)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel pets: %w", err)
	}

	query2 := `
	CREATE TABLE IF NOT EXISTS adoptions (
		id INT AUTO_INCREMENT PRIMARY KEY,
		pet_id INT NOT NULL,
		user_id INT NOT NULL,
		applicant_name VARCHAR(100) NOT NULL,
		applicant_phone VARCHAR(20) NOT NULL,
		applicant_email VARCHAR(100) NOT NULL,
		applicant_address TEXT NOT NULL,
		reason TEXT NOT NULL,
		status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
		reviewed_by INT NULL,
		reviewed_at TIMESTAMP NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE CASCADE,
		INDEX idx_adoptions_user_id (user_id),
		INDEX idx_adoptions_pet_id (pet_id),
		INDEX idx_adoptions_status (status)
	);
	`
	_, err = db.Exec(query2)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel adoptions: %w", err)
	}

	query3 := `
	CREATE TABLE IF NOT EXISTS notifications (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		title VARCHAR(200) NOT NULL,
		message TEXT NOT NULL,
		type ENUM('approved', 'rejected', 'info') NOT NULL,
		pet_name VARCHAR(100),
		is_read BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_notifications_user_id (user_id),
		INDEX idx_notifications_is_read (is_read)
	);
	`
	_, err = db.Exec(query3)
	if err != nil {
		return fmt.Errorf("gagal membuat tabel notifications: %w", err)
	}

	log.Println("Migrasi database pets berhasil")
	return nil
}
