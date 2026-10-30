package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iluha481/shieldgate/pkg/metrics"
)

func TestMetrics_CountsResponses(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		allowed uint64
		blocked uint64
		errors  uint64
	}{
		{name: "success", status: http.StatusOK, allowed: 1},
		{name: "rate limited", status: http.StatusTooManyRequests, blocked: 1},
		{name: "error", status: http.StatusInternalServerError, errors: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := metrics.NewCollectorManager()
			handler := Metrics(manager, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))

			// extractServicePrefix uses the request host as the metric key.
			req := httptest.NewRequest(http.MethodGet, "http://svc1.example.com/api/users", nil)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			col := manager.GetOrCreate("/svc1.example.com")
			if got := col.TotalRequests.Load(); got != 1 {
				t.Errorf("TotalRequests = %d, want 1", got)
			}
			if got := col.AllowedRequests.Load(); got != tt.allowed {
				t.Errorf("AllowedRequests = %d, want %d", got, tt.allowed)
			}
			if got := col.BlockedRequests.Load(); got != tt.blocked {
				t.Errorf("BlockedRequests = %d, want %d", got, tt.blocked)
			}
			if got := col.ErrorsRequests.Load(); got != tt.errors {
				t.Errorf("ErrorsRequests = %d, want %d", got, tt.errors)
			}
		})
	}
}

func TestMetrics_DefaultsToOKStatus(t *testing.T) {
	manager := metrics.NewCollectorManager()
	handler := Metrics(manager, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "http://svc.example.com/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	col := manager.GetOrCreate("/svc.example.com")
	if got := col.AllowedRequests.Load(); got != 1 {
		t.Errorf("AllowedRequests = %d, want 1", got)
	}
}

func TestExtractServicePrefix(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{host: "svc.example.com", want: "/svc.example.com"},
		{host: "", want: ""},
		{host: "/", want: ""},
	}

	for _, tt := range tests {
		if got := extractServicePrefix(tt.host); got != tt.want {
			t.Errorf("extractServicePrefix(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}
