package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/provenance-templates/url-shortener/db"
	"github.com/provenance-templates/url-shortener/middleware"
)

func GetAnalytics(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)
		slug := chi.URLParam(r, "slug")

		link, err := db.GetLink(database, slug)
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		if link.CreatedBy != userID {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}

		analytics, err := db.GetAnalytics(database, slug)
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analytics)
	}
}
