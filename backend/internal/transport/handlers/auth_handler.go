package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	tokenService        *service.TokenService
	userService         *service.UserService
	verificationService *service.VerificationService
	passwordRecovery    *service.PasswordRecoveryService
	cookieName          string
	cookieSecure        bool
}

// LoginRequest represents the expected payload for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the expected payload for registration requests.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	userService *service.UserService,
	tokenService *service.TokenService,
	verificationService *service.VerificationService,
	passwordRecovery *service.PasswordRecoveryService,
	cookieName string,
	cookieSecure bool,
) *AuthHandler {
	return &AuthHandler{
		userService:         userService,
		tokenService:        tokenService,
		verificationService: verificationService,
		passwordRecovery:    passwordRecovery,
		cookieName:          cookieName,
		cookieSecure:        cookieSecure,
	}
}

// Login authenticates a user and sets the authentication cookie.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, err := h.userService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "email and password are required",
			})
			return
		}

		if errors.Is(err, service.ErrInvalidEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid email address",
			})
			return
		}

		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return
		}

		if errors.Is(err, service.ErrUnverifiedEmail) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "unverified email",
			})
			return
		}

		slog.Error("failed to authenticate user", "error", err, "email", req.Email)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to authenticate user",
		})
		return
	}

	token, err := h.tokenService.GenerateToken(user.ID, user.Email, user.Username, user.Role)
	if err != nil {
		slog.Error("failed to generate auth token", "error", err, "user_id", user.ID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate auth token",
		})
		return
	}

	//nolint:gosec,nolintlint // Secure attribute is dynamically configured for local dev vs production
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.sameSiteMode(),
		MaxAge:   int(h.tokenService.TTL().Seconds()),
	})

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Register creates a new user account and sets the authentication cookie.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         "user",
	}

	if err := h.userService.Create(c.Request.Context(), user); err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "username, email and password are required",
			})
			return
		}

		if errors.Is(err, service.ErrInvalidEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid email address",
			})
			return
		}

		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already registered",
			})
			return
		}

		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "username or email already registered",
			})
			return
		}

		slog.Error("failed to register user", "error", err, "username", req.Username, "email", req.Email)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to register user",
		})
		return
	}

	if h.verificationService != nil {
		if err := h.verificationService.GenerateAndSendVerificationToken(c.Request.Context(), user); err != nil {
			slog.Error("failed to generate verification token", "error", err, "user_id", user.ID)
			// We still return 201 because the user is created, but maybe warn them
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully, please check your email",
		"user":    user,
	})
}

// VerifyEmail verifies a user's email address using a token.
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	if err := h.verificationService.VerifyToken(c.Request.Context(), tokenStr); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
			return
		}
		
		slog.Error("failed to verify email", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

// ResendVerificationRequest represents the expected payload for resending verification email.
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResendVerificationEmail resends the verification email to the user.
func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := h.verificationService.ResendVerificationEmail(c.Request.Context(), req.Email); err != nil {
		if err.Error() == "user is already verified" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user is already verified"})
			return
		}
		
		slog.Error("failed to resend verification email", "error", err, "email", req.Email)
		// Return 200 OK even if user doesn't exist to prevent email enumeration
		c.JSON(http.StatusOK, gin.H{"message": "If the email is registered and unverified, a new link has been sent."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email is registered and unverified, a new link has been sent."})
}

// Me returns the current authenticated user from the request context.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userIDString, _ := userID.(string)
	user, err := h.userService.GetByID(c.Request.Context(), userIDString)
	if err != nil {
		slog.Error("failed to retrieve authenticated user", "error", err, "user_id", userIDString)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve authenticated user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Logout clears the authentication cookie for the current client.
func (h *AuthHandler) Logout(c *gin.Context) {
	//nolint:gosec,nolintlint // Secure attribute is dynamically configured for local dev vs production
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: h.sameSiteMode(),
		MaxAge:   -1,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "session closed",
	})
}

func (h *AuthHandler) sameSiteMode() http.SameSite {
	if h.cookieSecure {
		return http.SameSiteNoneMode
	}

	return http.SameSiteLaxMode
}

// ForgotPasswordRequest payload for requesting password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword handles initiating a password reset.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := h.passwordRecovery.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		slog.Error("failed to process forgot password request", "error", err, "email", req.Email)
		// We still return 200 OK to prevent email enumeration
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email is registered, a password reset link has been sent."})
}

// ResetPasswordRequest payload for setting a new password.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword handles resetting the user's password using the token.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload, password is required"})
		return
	}

	if err := h.passwordRecovery.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidResetToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
			return
		}
		
		slog.Error("failed to reset password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}
