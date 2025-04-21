package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/user_model"
	"pvz/internal/services/user_service"
	"strings"
)

const (
	AuthUserKey  = "auth_user"
	AuthTokenKey = "auth_token"
)

type AuthMiddleware struct {
	userService user_service.IUserService
}

func NewAuthMiddleware(userService user_service.IUserService) *AuthMiddleware {
	return &AuthMiddleware{userService: userService}
}

func (m *AuthMiddleware) AuthMiddleware(c *gin.Context) {
	path := c.FullPath()
	if strings.Contains(path, "/dummyLogin") ||
		strings.Contains(path, "/login") ||
		strings.Contains(path, "/register") {
		c.Next()
		return
	}

	if _, exists := c.Get(generated.BearerAuthScopes); !exists {
		return
	}

	token, err := extractToken(c)
	if err != nil {
		log.Error().Err(err).Msg("Error extracting token")
		c.JSON(http.StatusUnauthorized, generated.Error{Message: "Unauthorized: " + err.Error()})
		c.Abort()
		return
	}

	user, err := m.userService.ValidateToken(c.Request.Context(), token)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusUnauthorized, generated.Error{Message: "Unauthorized: " + userErr.Error()})
		c.Abort()
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Error validating token")
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Auth error: " + err.Error()})
		c.Abort()
		return
	}

	c.Set(AuthUserKey, user)
	c.Set(AuthTokenKey, token)

	if !checkRole(c, user.Role, path) {
		log.Error().Msgf("Role %s not allowed", user.Role)
		c.JSON(http.StatusForbidden, generated.Error{Message: "Forbidden"})
		c.Abort()
		return
	}

	c.Next()
}

func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}

		return "", errors.New("invalid authorization header format")
	}

	token, err := c.Cookie("auth_token")
	if err == nil && token != "" {
		return token, nil
	}

	return "", errors.New("no authentication token found")
}

func checkRole(c *gin.Context, userRole user_model.UserRole, path string) bool {
	method := c.Request.Method

	if path == "/pvz" && method == http.MethodPost {
		return string(userRole) == string(generated.UserRoleModerator)
	}

	if strings.Contains(path, "/close_last_reception") && method == http.MethodPost ||
		strings.Contains(path, "/delete_last_product") && method == http.MethodPost ||
		strings.Contains(path, "/receptions") && method == http.MethodPost ||
		strings.Contains(path, "/products") && method == http.MethodPost {
		return string(userRole) == string(generated.UserRoleEmployee)
	}

	return true
}
