package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/provenance-templates/url-shortener/db"
	"github.com/provenance-templates/url-shortener/models"
)

func Redirect(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		link, err := db.GetLink(database, slug)
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}

		http.Redirect(w, r, link.Destination, http.StatusMovedPermanently)

		go recordClick(database, slug, r)
	}
}

func recordClick(database *sql.DB, slug string, r *http.Request) {
	ua := r.Header.Get("User-Agent")
	ref := r.Header.Get("Referer")
	ip := r.RemoteAddr

	click := &models.Click{
		Slug:      slug,
		UserAgent: nullableString(ua),
		Referer:   nullableString(ref),
		IPAddress: nullableString(ip),
	}
	_ = db.RecordClick(database, click)
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
