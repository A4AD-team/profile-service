package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/A4AD-team/profile-service/internal/domain"
	"github.com/A4AD-team/profile-service/internal/middleware"
	"github.com/A4AD-team/profile-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	svc      *service.ProfileService
	validate *validator.Validate
}

func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// GET /api/v1/profiles/{username}
func (h *ProfileHandler) GetProfileByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	profile, err := h.svc.GetProfileByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			respondError(w, http.StatusNotFound, "profile not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondJSON(w, http.StatusOK, profile)
}

// GET /api/v1/profiles/search?q=keyword&limit=20&offset=0
func (h *ProfileHandler) SearchProfiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := parseIntParam(r, "limit", 20)
	offset := parseIntParam(r, "offset", 0)

	profiles, err := h.svc.SearchProfiles(r.Context(), q, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"results": profiles,
		"count":   len(profiles),
	})
}

// GET /api/v1/profiles/me
func (h *ProfileHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r)
	profile, err := h.svc.GetMyProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			respondError(w, http.StatusNotFound, "profile not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondJSON(w, http.StatusOK, profile)
}

// PATCH /api/v1/profiles/me
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r)

	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	profile, err := h.svc.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			respondError(w, http.StatusNotFound, "profile not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	respondJSON(w, http.StatusOK, profile)
}

// GET /api/v1/profiles/me/stats
func (h *ProfileHandler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r)

	stats, err := h.svc.GetStats(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			respondError(w, http.StatusNotFound, "profile not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

// GET /api/v1/admin/profiles?limit=20&offset=0
func (h *ProfileHandler) ListAllProfiles(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 20)
	offset := parseIntParam(r, "offset", 0)

	profiles, err := h.svc.ListAllProfiles(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"results": profiles,
		"count":   len(profiles),
	})
}

// PATCH /api/v1/admin/profiles/{id}/reputation
func (h *ProfileHandler) UpdateReputation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	profileID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	var req domain.UpdateReputationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.svc.UpdateReputation(r.Context(), profileID, req.Delta); err != nil {
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"userId"   validate:"required,uuid"`
		Username string `json:"username" validate:"required,min=3,max=50"`
		Email    string `json:"email"    validate:"required,email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	profile, err := h.svc.CreateProfile(r.Context(), userID, req.Username, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			respondError(w, http.StatusConflict, "profile already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondJSON(w, http.StatusCreated, profile)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func parseIntParam(r *http.Request, key string, defaultVal int) int {
	val, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}
