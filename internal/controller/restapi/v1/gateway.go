package v1

import (
	"errors"
	"net/http"

	"shieldgate/internal/service"
	"shieldgate/pkg/proxy"
)

type GatewayController struct {
	gatewayService *service.GatewayService
	proxyManager   *proxy.ProxyManager
}

func NewGatewayController(
	gwService *service.GatewayService,
	pm *proxy.ProxyManager,
) *GatewayController {
	return &GatewayController{
		gatewayService: gwService,
		proxyManager:   pm,
	}
}

func (c *GatewayController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")

	result, err := c.gatewayService.HandleIncomingRequest(r.Context(), r.URL.Path, apiKey)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRouteNotFound):
			writeError(w, http.StatusNotFound, "route not found for this path")
		case errors.Is(err, service.ErrInvalidKey):
			writeError(w, http.StatusUnauthorized, "invalid or missing X-API-Key")
		case errors.Is(err, service.ErrRateExceeded):
			w.Header().Set("Retry-After", "1")
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		default:
			writeError(w, http.StatusInternalServerError, "internal gateway error")
		}
		return
	}

	// Проксируем дальше на целевой бэкенд
	if err := c.proxyManager.Forward(w, r, result.Service.TargetURL, result.Service.PathPrefix); err != nil {
		writeError(w, http.StatusBadGateway, "upstream unreachable")
	}
}
