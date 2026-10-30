package limiter

import (
	"context"
	"testing"
)

func TestAllow_NonPositiveLimitAlwaysAllows(t *testing.T) {
	// No Redis client is needed: non-positive limits short-circuit.
	l := NewRedisLimiter(nil)

	for _, limit := range []int{0, -1} {
		allowed, err := l.Allow(context.Background(), "key", limit)
		if err != nil {
			t.Fatalf("Allow(limit=%d) error = %v", limit, err)
		}
		if !allowed {
			t.Errorf("Allow(limit=%d) = false, want true", limit)
		}
	}
}
