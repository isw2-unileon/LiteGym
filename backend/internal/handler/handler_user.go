package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/repositories"
)

type UserHandler struct {
	repo *repositories.UserRepository
}

func NewUserHandler(repo *repositories.UserRepository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user model.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username, email and password are required",
		})
		return
	}

	if err := h.repo.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}