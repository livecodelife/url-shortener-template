package db

import (
	"database/sql"
	"errors"

	"github.com/provenance-templates/url-shortener/models"
)

var ErrNotFound = errors.New("not found")
var ErrSlugConflict = errors.New("slug conflict")

func GetLink(db *sql.DB, slug string) (*models.Link, error) {
	link := &models.Link{}
	err := db.QueryRow(
		`SELECT slug, destination, created_at, created_by FROM links WHERE slug = $1`,
		slug,
	).Scan(&link.Slug, &link.Destination, &link.CreatedAt, &link.CreatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return link, err
}

func SlugExists(db *sql.DB, slug string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM links WHERE slug = $1`, slug).Scan(&count)
	return count > 0, err
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM clicks WHERE slug = $1`, slug)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM links WHERE slug = $1`, slug)
	if err != nil {
		return err
	}

	return tx.Commit()
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

