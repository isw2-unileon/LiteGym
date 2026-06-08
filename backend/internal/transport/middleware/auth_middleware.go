package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

// Context keys used by the authentication middleware to store user values in the request context.
//
// These keys are placed into the Gin context by the middleware and can be read by
// handlers to obtain the authenticated user's ID, email, username and role.
const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
	ContextUsernameKey  = "username"
	ContextUserRoleKey  = "user_role"
)

// AuthMiddleware provides HTTP middleware for authentication.
//
// It uses a TokenService to parse and validate authentication tokens stored in a cookie,
// and injects the parsed claims into the request context for downstream handlers.
type AuthMiddleware struct {
	tokenService *service.TokenService
	userService  *service.UserService
	cookieName   string
}

// NewAuthMiddleware constructs a new AuthMiddleware.
//
// tokenService is used to parse/verify tokens and cookieName specifies which cookie
// contains the authentication token.
func NewAuthMiddleware(tokenService *service.TokenService, cookieName string, userService ...*service.UserService) *AuthMiddleware {
	var resolvedUserService *service.UserService
	if len(userService) > 0 {
		resolvedUserService = userService[0]
	}

	return &AuthMiddleware{
		tokenService: tokenService,
		userService:  resolvedUserService,
		cookieName:   cookieName,
	}
}

// RequireAuth returns a Gin middleware handler that enforces authentication.
//
// The returned handler reads the configured cookie, validates the token using the
// TokenService, and on success stores the claims (subject, email, username, role) in the
// request context.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(m.cookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: missing cookie"})
			return
		}

		claims, err := m.tokenService.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: invalid or expired token"})
			return
		}

		if m.userService != nil {
			if _, err := m.userService.GetByID(c.Request.Context(), claims.Subject); err != nil {
				if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidUserInput) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: user not found"})
					return
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to validate authenticated user"})
				return
			}
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextUserRoleKey, claims.Role)
		c.Next()
	}
}
