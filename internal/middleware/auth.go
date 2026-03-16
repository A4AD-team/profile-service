package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "userID"
	ContextKeyRoles  contextKey = "roles"
)

// JWTAuth validates the Bearer token and injects userID + roles into context.
//
// JWT claims structure (from auth-service JwtService.issueAccessToken):
//   - sub         → user UUID string
//   - email       → user email
//   - roles       → []string e.g. ["USER"], ["ADMIN"]
//   - permissions → []string
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				respondUnauthorized(w, "missing authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				// auth-service uses HS256
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				respondUnauthorized(w, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondUnauthorized(w, "invalid claims")
				return
			}

			// Extract sub → userID
			sub, _ := claims["sub"].(string)
			userID, err := uuid.Parse(sub)
			if err != nil {
				respondUnauthorized(w, "invalid subject in token")
				return
			}

			// Extract roles — auth-service stores them as []interface{}
			roles := parseRoles(claims["roles"])

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyRoles, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin checks that the authenticated user has "ADMIN" role.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles, _ := r.Context().Value(ContextKeyRoles).([]string)
		for _, role := range roles {
			if role == "ADMIN" {
				next.ServeHTTP(w, r)
				return
			}
		}
		respondForbidden(w)
	})
}

// UserIDFromCtx extracts the userID set by JWTAuth middleware.
func UserIDFromCtx(r *http.Request) uuid.UUID {
	id, _ := r.Context().Value(ContextKeyUserID).(uuid.UUID)
	return id
}

// parseRoles converts []interface{} from JWT claims to []string.
func parseRoles(raw any) []string {
	rawSlice, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	roles := make([]string, 0, len(rawSlice))
	for _, v := range rawSlice {
		if s, ok := v.(string); ok {
			roles = append(roles, s)
		}
	}
	return roles
}

func respondUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
}

func InternalSecret(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Internal-Secret") != secret {
				respondForbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
