package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/labstack/echo/v4"
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

// Auth middleware for protected routes (http.Handler version)
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

// AuthEcho middleware for Echo framework protected routes
func (m *AuthMiddleware) AuthEcho(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cctx := c.Request().Context()
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			logger.Error(cctx, "No authorization header found")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Error(cctx, "Invalid authorization header")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
		}

		token := parts[1]

		claims, err := utils.ValidateJWT(token, m.jwtSecret)
		if err != nil {
			logger.Error(cctx, "Invalid token")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": fmt.Sprintf("invalid token: %v", err)})
		}

		key := fmt.Sprintf("jwt:user:%d", claims.UserID)
		storedToken, err := m.redis.Get(c.Request().Context(), key).Result()
		if err == redis.Nil {
			logger.Error(cctx, "Token not found")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token expired or logged out"})
		}
		if err != nil {
			logger.Error(cctx, "Invalid token")
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to verify token"})
		}
		if storedToken != token {
			logger.Error(cctx, "Invalid token")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token mismatch"})
		}

		// Set values in Echo context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("driver_id", claims.UserID)

		fmt.Println("user id from JWT:", claims.UserID, " role: ", claims.Role)
		return next(c)
	}
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

// Echo-specific helper functions
func GetUserIDFromEcho(c echo.Context) (int64, bool) {
	userID, ok := c.Get("user_id").(int64)
	return userID, ok
}

func GetUserRoleFromEcho(c echo.Context) (string, bool) {
	role, ok := c.Get("user_role").(string)
	return role, ok
}

func GetDriverIDFromEcho(c echo.Context) (int64, bool) {
	driverID, ok := c.Get("driver_id").(int64)
	return driverID, ok
}
