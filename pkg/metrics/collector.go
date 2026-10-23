package metrics

import (
	"sync/atomic"
)

type Collector struct {
	TotalRequests   atomic.Uint64
	AllowedRequests atomic.Uint64
	BlockedRequests atomic.Uint64
	ErrorsRequests  atomic.Uint64
}

func NewCollector() *Collector {
	return &Collector{}
}

type Metric struct {
	ServiceID int    `json:"service_id"`
	RPS       uint64 `json:"rps"`
	Allowed   uint64 `json:"allowed"`
	Blocked   uint64 `json:"blocked"`
	Errors    uint64 `json:"errors"`
}
