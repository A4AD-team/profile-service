package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "0.1.0",
	})
}

// GET /health/live — liveness probe, always 200 if process is running
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// GET /health/ready — readiness probe, checks DB connection
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "healthy"
	status := http.StatusOK

	if err := h.pool.Ping(ctx); err != nil {
		dbStatus = "unhealthy"
		status = http.StatusServiceUnavailable
	}

	respondJSON(w, status, map[string]any{
		"status":    dbStatus,
		"timestamp": time.Now().UTC(),
		"checks": map[string]string{
			"database": dbStatus,
		},
	})
}
