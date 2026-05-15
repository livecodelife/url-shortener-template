package db

import (
	"database/sql"
	"errors"

	"github.com/provenance-templates/identity/models"
)

var ErrNotFound = errors.New("not found")
var ErrEmailExists = errors.New("email already exists")

func GetUserByEmail(database *sql.DB, email string) (*models.UserWithPassword, error) {
	var idVal, emailVal, passwordVal sql.NullString
	err := database.QueryRow(
		`SELECT id, email, password FROM users WHERE email = $1`,
		email,
	).Scan(&idVal, &emailVal, &passwordVal)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !idVal.Valid) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &models.UserWithPassword{
		ID:       idVal.String,
		Email:    emailVal.String,
		Password: passwordVal.String,
	}, nil
}

func EmailExists(database *sql.DB, email string) (bool, error) {
	var s sql.NullString
	err := database.QueryRow(`SELECT id FROM users WHERE email = $1 LIMIT 1`, email).Scan(&s)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return s.Valid, nil
}

func CreateUser(database *sql.DB, id, email, passwordHash string) error {
	_, err := database.Exec(
		`INSERT INTO users (id, email, password) VALUES ($1, $2, $3)`,
		id, email, passwordHash,
	)
	return err
}

func CreateToken(database *sql.DB, token, userID string) error {
	_, err := database.Exec(
		`INSERT INTO tokens (token, user_id) VALUES ($1, $2)`,
		token, userID,
	)
	return err
}

func GetUserIDByToken(database *sql.DB, token string) (string, error) {
	var userID sql.NullString
	err := database.QueryRow(
		`SELECT user_id FROM tokens WHERE token = $1`,
		token,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !userID.Valid) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return userID.String, nil
}
