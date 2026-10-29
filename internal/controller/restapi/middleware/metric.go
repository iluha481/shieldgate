package middleware

import (
	"net/http"
	"strings"

	"github.com/iluha481/shieldgate/pkg/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func Metrics(cm *metrics.CollectorManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(recorder, r)
		prefix := extractServicePrefix(r.URL.Host)
		collector := cm.GetOrCreate(prefix)
		collector.TotalRequests.Add(1)
		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			collector.AllowedRequests.Add(1)
		} else if recorder.statusCode == http.StatusTooManyRequests {
			collector.BlockedRequests.Add(1)
		} else {
			collector.ErrorsRequests.Add(1)
		}
	})
}

func extractServicePrefix(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return "/" + parts[0]
}
