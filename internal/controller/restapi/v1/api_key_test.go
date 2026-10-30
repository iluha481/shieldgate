package v1

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iluha481/shieldgate/internal/domain"
	"github.com/iluha481/shieldgate/internal/service"
)

func TestAPIKeyController_Generate(t *testing.T) {
	svcRepo := &fakeServiceRepo{svc: &domain.Service{ID: 5, UserID: 1}}
	keyRepo := &fakeAPIKeyRepo{}
	ctrl := NewAPIKeyController(service.NewRouteService(svcRepo, keyRepo))

	req := withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services/5/keys",
		strings.NewReader(`{"name":"ci-key"}`)), "1")
	req.SetPathValue("id", "5")
	rec := httptest.NewRecorder()

	ctrl.Generate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	body := decodeBody(t, rec.Body.Bytes())
	plain, ok := body["plain_key"].(string)
	if !ok || !strings.HasPrefix(plain, "sg_live_") {
		t.Fatalf("plain_key = %v, want sg_live_ prefix", body["plain_key"])
	}
	apiKey, ok := body["api_key"].(map[string]any)
	if !ok {
		t.Fatalf("api_key is missing or has wrong type: %v", body["api_key"])
	}
	if apiKey["name"] != "ci-key" || apiKey["service_id"] != float64(5) {
		t.Errorf("unexpected api key: %v", apiKey)
	}
	if keyRepo.created == nil || keyRepo.created.KeyHash == "" {
		t.Error("api key was not persisted")
	}
}

func TestAPIKeyController_Generate_Unauthorized(t *testing.T) {
	ctrl := NewAPIKeyController(service.NewRouteService(&fakeServiceRepo{}, &fakeAPIKeyRepo{}))
	rec := httptest.NewRecorder()

	ctrl.Generate(rec, httptest.NewRequest(http.MethodPost, "/api/v1/services/5/keys", strings.NewReader(`{}`)))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAPIKeyController_Generate_InvalidID(t *testing.T) {
	ctrl := NewAPIKeyController(service.NewRouteService(&fakeServiceRepo{}, &fakeAPIKeyRepo{}))
	req := withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services/x/keys", strings.NewReader(`{}`)), "1")
	req.SetPathValue("id", "x")
	rec := httptest.NewRecorder()

	ctrl.Generate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAPIKeyController_Generate_InvalidJSON(t *testing.T) {
	ctrl := NewAPIKeyController(service.NewRouteService(&fakeServiceRepo{}, &fakeAPIKeyRepo{}))
	req := withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services/5/keys", strings.NewReader(`{`)), "1")
	req.SetPathValue("id", "5")
	rec := httptest.NewRecorder()

	ctrl.Generate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAPIKeyController_Generate_ServiceError(t *testing.T) {
	ctrl := NewAPIKeyController(service.NewRouteService(
		&fakeServiceRepo{getByIDErr: errInternal}, &fakeAPIKeyRepo{}))
	req := withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services/5/keys",
		strings.NewReader(`{"name":"key"}`)), "1")
	req.SetPathValue("id", "5")
	rec := httptest.NewRecorder()

	ctrl.Generate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
