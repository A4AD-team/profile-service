package handler

import (
	"time"

	"github.com/A4AD-team/profile-service/internal/config"
	"github.com/A4AD-team/profile-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(
	r *chi.Mux,
	cfg *config.Config,
	profileHandler *ProfileHandler,
	healthHandler *HealthHandler,
) {
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	r.Get("/health", healthHandler.Health)
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)
	r.Handle("/metrics", promhttp.Handler())

	r.With(middleware.InternalSecret(cfg.Auth.InternalSecret)).
		Post("/internal/profiles", profileHandler.CreateProfile)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))

		r.Get("/profiles/me", profileHandler.GetMyProfile)
		r.Patch("/profiles/me", profileHandler.UpdateProfile)
		r.Get("/profiles/me/stats", profileHandler.GetMyStats)
		r.Get("/profiles/search", profileHandler.SearchProfiles)
		r.Get("/profiles/{username}", profileHandler.GetProfileByUsername)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/admin/profiles", profileHandler.ListAllProfiles)
			r.Patch("/admin/profiles/{id}/reputation", profileHandler.UpdateReputation)
		})
	})
}
