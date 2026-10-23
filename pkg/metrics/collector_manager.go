package metrics

import "sync"

type CollectorManager struct {
	services sync.Map
}

func NewCollectorManager() *CollectorManager {
	return &CollectorManager{}
}

func (c *CollectorManager) GetOrCreate(path string) *Collector {
	if col, ok := c.services.Load(path); ok {
		return col.(*Collector)
	}
	col, _ := c.services.LoadOrStore(path, NewCollector())
	return col.(*Collector)
}

func (c *CollectorManager) Collect() []Metric {
	metrics := make([]Metric, 0)
	c.services.Range(func(key, value interface{}) bool {
		serviceID := key.(int)
		collector := value.(*Collector)
		total := collector.TotalRequests.Swap(0)
		allowed := collector.AllowedRequests.Swap(0)
		blocked := collector.BlockedRequests.Swap(0)
		errors := collector.ErrorsRequests.Swap(0)
		metrics = append(metrics, Metric{
			ServiceID: serviceID,
			RPS:       total,
			Allowed:   allowed,
			Blocked:   blocked,
			Errors:    errors,
		})
		return true
	})
	return metrics
}
