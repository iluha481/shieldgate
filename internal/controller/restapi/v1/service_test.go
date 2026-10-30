package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iluha481/shieldgate/internal/domain"
	"github.com/iluha481/shieldgate/internal/service"
)

func TestServiceController_Create(t *testing.T) {
	repo := &fakeServiceRepo{}
	ctrl := NewServiceController(service.NewRouteService(repo, &fakeAPIKeyRepo{}))

	req := withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services",
		bytes.NewBufferString(`{"name":"My API","path_prefix":"/api","target_url":"http://upstream"}`)), "1")
	rec := httptest.NewRecorder()

	ctrl.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}

	body := decodeBody(t, rec.Body.Bytes())
	if body["name"] != "My API" || body["path_prefix"] != "/api" {
		t.Errorf("unexpected response: %v", body)
	}
	if body["rate_limit_rps"] != float64(10) || body["burst_capacity"] != float64(20) {
		t.Errorf("defaults not applied: %v", body)
	}
	if repo.created == nil || repo.created.UserID != 1 {
		t.Errorf("service was not created for user 1: %+v", repo.created)
	}
}

func TestServiceController_Create_Unauthorized(t *testing.T) {
	ctrl := NewServiceController(service.NewRouteService(&fakeServiceRepo{}, &fakeAPIKeyRepo{}))
	rec := httptest.NewRecorder()

	ctrl.Create(rec, httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewBufferString(`{}`)))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestServiceController_Create_InvalidJSON(t *testing.T) {
	ctrl := NewServiceController(service.NewRouteService(&fakeServiceRepo{}, &fakeAPIKeyRepo{}))
	rec := httptest.NewRecorder()

	ctrl.Create(rec, withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewBufferString(`{`)), "1"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServiceController_Create_RepoError(t *testing.T) {
	repo := &fakeServiceRepo{createErr: domain.ErrPathPrefixAlreadyExists}
	ctrl := NewServiceController(service.NewRouteService(repo, &fakeAPIKeyRepo{}))
	rec := httptest.NewRecorder()

	ctrl.Create(rec, withUserID(t, httptest.NewRequest(http.MethodPost, "/api/v1/services",
		bytes.NewBufferString(`{"name":"My API","path_prefix":"/api","target_url":"http://upstream"}`)), "1"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServiceController_GetAll(t *testing.T) {
	repo := &fakeServiceRepo{services: []domain.Service{{ID: 1, Name: "one"}, {ID: 2, Name: "two"}}}
	ctrl := NewServiceController(service.NewRouteService(repo, &fakeAPIKeyRepo{}))

	rec := httptest.NewRecorder()
	ctrl.GetAll(rec, withUserID(t, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil), "1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var services []domain.Service
	if err := json.Unmarshal(rec.Body.Bytes(), &services); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(services) != 2 {
		t.Errorf("got %d services, want 2", len(services))
	}
}

func TestServiceController_GetAll_Errors(t *testing.T) {
	ctrl := NewServiceController(service.NewRouteService(&fakeServiceRepo{getByUserErr: errInternal}, &fakeAPIKeyRepo{}))

	t.Run("unauthorized", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctrl.GetAll(rec, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ctrl.GetAll(rec, withUserID(t, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil), "1"))
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}
