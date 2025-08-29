package handlers

import (
	"log/slog"
	"net/http"
	"to-do-list/internal/api/types"
	"to-do-list/internal/auth"
	"to-do-list/internal/models"
	"to-do-list/internal/repository"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type UserHandler struct {
	repo         repository.UserRepository
	tokenManager *auth.TokenManager
}

func NewUserHandler(repo repository.UserRepository, tm *auth.TokenManager) *UserHandler {
	return &UserHandler{
		repo:         repo,
		tokenManager: tm,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req types.RegisterRequest
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		slog.Error("Failed to decode request body", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})

		return
	}

	err = validate.Struct(req)
	if err != nil {
		slog.Error("Validation failed", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Validation failed: " + err.Error()})

		return
	}

	user, err := models.NewUser(req.Username, req.Email, req.Password)
	if err != nil {
		slog.Error("Failed to create new user model", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": err.Error()})

		return
	}

	err = h.repo.CreateUser(r.Context(), user)
	if err != nil {
		slog.Error("Failed to create user in repository", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})

		return
	}

	token, err := h.tokenManager.GenerateToken(*user)
	if err != nil {
		slog.Error("Failed to generate token", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})

		return
	}

	resp := types.AuthResponse{
		User: types.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.LoginRequest
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		slog.Error("Failed to decode request body", slog.Any("error", err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})

		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Error("User not found", slog.Any("error", err))

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Invalid credentials"})

		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(req.Password))
	if err != nil {
		slog.Error("Invalid password", slog.Any("error", err))

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Invalid credentials"})

		return
	}

	token, err := h.tokenManager.GenerateToken(user)
	if err != nil {
		slog.Error("Failed to generate token", slog.Any("error", err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})

		return
	}

	resp := types.AuthResponse{
		User: types.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
