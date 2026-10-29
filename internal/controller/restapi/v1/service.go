package v1

import (
	"encoding/json"
	"net/http"

	"github.com/iluha481/shieldgate/internal/controller/restapi/middleware"
	"github.com/iluha481/shieldgate/internal/service"
)

type ServiceController struct {
	routeService *service.RouteService
}

func NewServiceController(routeService *service.RouteService) *ServiceController {
	return &ServiceController{routeService: routeService}
}

type createServiceDTO struct {
	Name          string `json:"name"`
	PathPrefix    string `json:"path_prefix"`
	TargetURL     string `json:"target_url"`
	RateLimitRPS  int    `json:"rate_limit_rps"`
	BurstCapacity int    `json:"burst_capacity"`
}

func (c *ServiceController) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var dto createServiceDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	svc, err := c.routeService.CreateService(
		r.Context(),
		userID,
		dto.Name,
		dto.PathPrefix,
		dto.TargetURL,
		dto.RateLimitRPS,
		dto.BurstCapacity,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, svc)
}

func (c *ServiceController) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	services, err := c.routeService.GetUserServices(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get services")
		return
	}

	writeJSON(w, http.StatusOK, services)
}
