package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/provenance-templates/url-shortener/models"
)

var ErrNotFound = errors.New("not found")
var ErrSlugConflict = errors.New("slug conflict")

func GetLink(db *sql.DB, slug string) (*models.Link, error) {
	var (
		slugVal, destVal, createdBy sql.NullString
		createdAt                   sql.NullTime
	)
	err := db.QueryRow(
		`SELECT slug, destination, created_at, created_by FROM links WHERE slug = $1`,
		slug,
	).Scan(&slugVal, &destVal, &createdAt, &createdBy)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !slugVal.Valid) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t := time.Time{}
	if createdAt.Valid {
		t = createdAt.Time
	}
	return &models.Link{
		Slug:        slugVal.String,
		Destination: destVal.String,
		CreatedAt:   t,
		CreatedBy:   createdBy.String,
	}, nil
}

func SlugExists(db *sql.DB, slug string) (bool, error) {
	var s sql.NullString
	err := db.QueryRow(`SELECT slug FROM links WHERE slug = $1 LIMIT 1`, slug).Scan(&s)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return s.Valid, nil
}

func CreateLink(db *sql.DB, link *models.Link) error {
	exists, err := SlugExists(db, link.Slug)
	if err != nil {
		return err
	}
	if exists {
		return ErrSlugConflict
	}
	_, err = db.Exec(
		`INSERT INTO links (slug, destination, created_by) VALUES ($1, $2, $3)`,
		link.Slug, link.Destination, link.CreatedBy,
	)
	return err
}

func DeleteLink(db *sql.DB, slug string) error {
	_, err := db.Exec(`DELETE FROM links WHERE slug = $1`, slug)
	return err
}

func ListLinksByUser(db *sql.DB, userID string) ([]models.Link, error) {
	rows, err := db.Query(
		`SELECT slug, destination, created_at, created_by FROM links WHERE created_by = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []models.Link
	for rows.Next() {
		var l models.Link
		if err := rows.Scan(&l.Slug, &l.Destination, &l.CreatedAt, &l.CreatedBy); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if links == nil {
		links = []models.Link{}
	}
	return links, rows.Err()
}

