package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/provenance-templates/url-shortener/db"
	"github.com/provenance-templates/url-shortener/handlers"
	authmw "github.com/provenance-templates/url-shortener/middleware"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/{slug}", handlers.Redirect(database))

	r.Group(func(r chi.Router) {
		r.Use(authmw.RequireAuth)

		r.Post("/links", handlers.CreateLink(database))
		r.Get("/links", handlers.ListLinks(database))
		r.Delete("/links/{slug}", handlers.DeleteLink(database))
		r.Get("/links/{slug}/analytics", handlers.GetAnalytics(database))
	})

	log.Println("starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
