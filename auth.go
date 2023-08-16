package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"passwordHash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func CreateUser(db *sql.DB, username string, password, role string) (int, error) {
	username = strings.TrimSpace(username)
	if len(username) < 1 || len(username) > 64 {
		return 0, errors.New("username either too long or too short")
	}

	normalizedUsername := strings.ToUpper(username)
	passwordHash := hashPassword(password)

	res, err := db.Exec(`INSERT INTO users (username, normalized_username, password_hash, role) VALUES ($1, $2, $3, $4)`, username, normalizedUsername, passwordHash, role)
	if err != nil {
		return 0, err
	}

	return id32(res)
}

func CreateUserSession(db *sql.DB, userId int, userAgent, ipAddress string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to generate session token: %e", err)
	}
	sessionId := base64.URLEncoding.EncodeToString(b)
	_, err := db.Exec(`INSERT INTO user_sessions (session_id, user_id, user_agent, ip_address) VALUES ($1, $2, $3, $4)`, sessionId, userId, userAgent, ipAddress)
	return sessionId, err
}

func FindUser(db *sql.DB, username string) (user User, err error) {
	username = strings.ToUpper(strings.TrimSpace(username))
	err = db.QueryRow(`SELECT id, username, password_hash, role, created_at  FROM users WHERE normalized_username = $1`, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	return
}

func TryLogin(db *sql.DB, username, password string) (User, error) {
	user, err := FindUser(db, username)
	if err != nil {
		return User{}, err
	}

	if comparePasswordHash(user.PasswordHash, password) {
		return user, nil
	} else {
		return User{}, errors.New("invalid password")
	}
}

func hashPassword(pwd string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func comparePasswordHash(hash, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
