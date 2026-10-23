package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"shieldgate/internal/controller/restapi/middleware"
	"shieldgate/internal/service"
)

type APIKeyController struct {
	routeService *service.RouteService
}

func NewAPIKeyController(routeService *service.RouteService) *APIKeyController {
	return &APIKeyController{routeService: routeService}
}

type generateKeyDTO struct {
	Name string `json:"name"`
}

// Generate — POST /api/v1/services/{id}/keys
func (c *APIKeyController) Generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Go 1.22: Достаем {id} из URL
	serviceID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}

	var dto generateKeyDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	plainKey, apiKey, err := c.routeService.GenerateAPIKey(r.Context(), userID, serviceID, dto.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Возвращаем plain_key пользователю (СТРОГО ОДИН РАЗ!)
	writeJSON(w, http.StatusCreated, map[string]any{
		"api_key":   apiKey,
		"plain_key": plainKey,
	})
}
