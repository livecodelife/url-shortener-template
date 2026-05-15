package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/provenance-templates/url-shortener/db"
	"github.com/provenance-templates/url-shortener/middleware"
	"github.com/provenance-templates/url-shortener/models"
)

type createLinkRequest struct {
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
}

func CreateLink(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		var req createLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}

		slug := req.Slug
		userProvidedSlug := slug != ""
		if !userProvidedSlug {
			var err error
			slug, err = generateSlug()
			if err != nil {
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}
		}

		link := &models.Link{
			Slug:        slug,
			Destination: req.Destination,
			CreatedBy:   userID,
		}

		if err := db.CreateLink(database, link); err != nil {
			if errors.Is(err, db.ErrSlugConflict) {
				if userProvidedSlug {
					http.Error(w, `{"error":"slug already exists"}`, http.StatusConflict)
					return
				}
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		created, err := db.GetLink(database, slug)
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

func DeleteLink(database *sql.DB) http.HandlerFunc {
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

		if err := db.DeleteLink(database, slug); err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListLinks(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r)

		links, err := db.ListLinksByUser(database, userID)
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)
	}
}

func generateSlug() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
