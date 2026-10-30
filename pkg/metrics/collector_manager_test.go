package metrics

import "testing"

func TestCollectorManager_GetOrCreateReturnsSameInstance(t *testing.T) {
	cm := NewCollectorManager()

	first := cm.GetOrCreate("/api")
	second := cm.GetOrCreate("/api")

	if first != second {
		t.Fatal("GetOrCreate() returned different collectors for the same path")
	}
	if other := cm.GetOrCreate("/other"); other == first {
		t.Fatal("GetOrCreate() returned the same collector for a different path")
	}
}

func TestCollectorManager_Collect(t *testing.T) {
	cm := NewCollectorManager()
	col := NewCollector()
	col.TotalRequests.Add(6)
	col.AllowedRequests.Add(3)
	col.BlockedRequests.Add(2)
	col.ErrorsRequests.Add(1)
	cm.services.Store(7, col)

	got := cm.Collect()
	if len(got) != 1 {
		t.Fatalf("Collect() returned %d metrics, want 1", len(got))
	}

	want := Metric{ServiceID: 7, RPS: 6, Allowed: 3, Blocked: 2, Errors: 1}
	if got[0] != want {
		t.Errorf("Collect() = %+v, want %+v", got[0], want)
	}

	// Collect is a snapshot: counters are reset afterwards.
	again := cm.Collect()
	if len(again) != 1 || again[0] != (Metric{ServiceID: 7}) {
		t.Errorf("second Collect() = %+v, want zeroed metric", again)
	}
}

func TestCollectorManager_CollectEmpty(t *testing.T) {
	got := NewCollectorManager().Collect()
	if len(got) != 0 {
		t.Errorf("Collect() on empty manager returned %d metrics, want 0", len(got))
	}
}
