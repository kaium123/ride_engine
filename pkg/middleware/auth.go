package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/redis/go-redis/v9"
	"vcs.technonext.com/carrybee/ride_engine/pkg/utils"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
	DriverIdKey contextKey = "driver_id"
)

type AuthMiddleware struct {
	redis     *redis.Client
	jwtSecret string
}

func NewAuthMiddleware(redisClient *redis.Client, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		redis:     redisClient,
		jwtSecret: jwtSecret,
	}
}

// Auth middleware for protected routes
func (m *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cctx := r.Context()
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Error(cctx, "No authorization header found")
			sendError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Error(cctx, "Invalid authorization header")
			sendError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		token := parts[1]

		claims, err := utils.ValidateJWT(token, m.jwtSecret)
		if err != nil {
			logger.Error(cctx, "Invalid token")
			sendError(w, http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
			return
		}

		key := fmt.Sprintf("jwt:user:%d", claims.UserID)
		storedToken, err := m.redis.Get(r.Context(), key).Result()
		if err == redis.Nil {
			logger.Error(cctx, "Token not found")
			sendError(w, http.StatusUnauthorized, "token expired or logged out")
			return
		}
		if err != nil {
			logger.Error(cctx, "Invalid token")
			sendError(w, http.StatusInternalServerError, "failed to verify token")
			return
		}
		if storedToken != token {
			logger.Error(cctx, "Invalid token")
			sendError(w, http.StatusUnauthorized, "token mismatch")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, DriverIdKey, claims.UserID)

		fmt.Println("driver id from JWT:", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware to check user role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cctx := r.Context()
			userRole := r.Context().Value(UserRoleKey)
			if userRole == nil {
				logger.Error(cctx, "User role not found")
				sendError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if userRole.(string) != role {
				logger.Error(cctx, "User role mismatch")
				sendError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// Helper function to send JSON error response
func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func GetDriverID(ctx context.Context) (int64, bool) {
	driverID, ok := ctx.Value(DriverIdKey).(int64)
	return driverID, ok
}
