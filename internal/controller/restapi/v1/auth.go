package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"shieldgate/internal/domain"
	"shieldgate/internal/service"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

type authDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var dto authDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	user, err := c.authService.Register(r.Context(), dto.Email, dto.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, "user with this email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var dto authDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	token, err := c.authService.Login(r.Context(), dto.Email, dto.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
