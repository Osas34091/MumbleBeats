package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a dashboard user
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Session represents an active login session
type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
}

// HasAdmin check if any user exists
func HasAdmin() (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateUser hashes the password and saves the user
func CreateUser(username, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	res, err := DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, string(hash))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// VerifyLogin checks credentials and returns the user ID if valid
func VerifyLogin(username, password string) (int64, error) {
	var id int64
	var hash string
	err := DB.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&id, &hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("credenciales inválidas")
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return 0, errors.New("credenciales inválidas")
	}
	return id, nil
}

// CreateSession creates a new cryptographically secure session token
func CreateSession(userID int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	
	// Set expiration to 30 days
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	
	_, err := DB.Exec("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)", token, userID, expiresAt)
	if err != nil {
		return "", err
	}
	
	return token, nil
}

// GetUserFromSession validates the token and returns the username
func GetUserFromSession(token string) (string, error) {
	var username string
	var expiresAt time.Time
	
	err := DB.QueryRow(`
		SELECT u.username, s.expires_at 
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = ?
	`, token).Scan(&username, &expiresAt)
	
	if err != nil {
		return "", errors.New("sesión no válida")
	}
	
	if time.Now().After(expiresAt) {
		DeleteSession(token)
		return "", errors.New("sesión expirada")
	}
	
	return username, nil
}

// DeleteSession removes a session
func DeleteSession(token string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
