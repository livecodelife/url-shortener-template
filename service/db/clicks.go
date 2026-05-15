package db

import (
	"database/sql"

	"github.com/provenance-templates/url-shortener/models"
)

type AnalyticsResult struct {
	TotalClicks   int            `json:"total_clicks"`
	ClicksLast24h int            `json:"clicks_last_24h"`
	RecentClicks  []models.Click `json:"recent_clicks"`
}

func RecordClick(db *sql.DB, click *models.Click) error {
	_, err := db.Exec(
		`INSERT INTO clicks (slug, user_agent, referer, ip_address) VALUES ($1, $2, $3, $4)`,
		click.Slug, click.UserAgent, click.Referer, click.IPAddress,
	)
	return err
}

func GetAnalytics(db *sql.DB, slug string) (*AnalyticsResult, error) {
	result := &AnalyticsResult{}

	err := db.QueryRow(`SELECT COUNT(*) FROM clicks WHERE slug = $1`, slug).Scan(&result.TotalClicks)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(
		`SELECT COUNT(*) FROM clicks WHERE slug = $1 AND clicked_at > NOW() - INTERVAL '24 hours'`,
		slug,
	).Scan(&result.ClicksLast24h)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT slug, clicked_at, user_agent, referer, ip_address FROM clicks WHERE slug = $1 ORDER BY clicked_at DESC LIMIT 10`,
		slug,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result.RecentClicks = []models.Click{}
	for rows.Next() {
		var c models.Click
		if err := rows.Scan(&c.Slug, &c.ClickedAt, &c.UserAgent, &c.Referer, &c.IPAddress); err != nil {
			return nil, err
		}
		result.RecentClicks = append(result.RecentClicks, c)
	}

	return result, rows.Err()
}
