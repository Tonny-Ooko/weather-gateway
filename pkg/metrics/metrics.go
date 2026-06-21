package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

type ServiceMetrics struct {
	TotalRequests        int64 `json:"total_requests"`
	CacheHits            int64 `json:"cache_hits"`
	CacheMisses          int64 `json:"cache_misses"`
	UpstreamFailures     int64 `json:"upstream_failures"`
	RateLimitedRequests  int64 `json:"rate_limited_requests"`
	CircuitBreakerTrips  int64 `json:"circuit_breaker_trips"`
}

var GlobalMetrics = &ServiceMetrics{}

func IncrementTotalRequests()       { atomic.AddInt64(&GlobalMetrics.TotalRequests, 1) }
func IncrementCacheHits()           { atomic.AddInt64(&GlobalMetrics.CacheHits, 1) }
func IncrementCacheMisses()         { atomic.AddInt64(&GlobalMetrics.CacheMisses, 1) }
func IncrementUpstreamFailures()    { atomic.AddInt64(&GlobalMetrics.UpstreamFailures, 1) }
func IncrementRateLimitedRequests() { atomic.AddInt64(&GlobalMetrics.RateLimitedRequests, 1) }
func IncrementCircuitBreakerTrips() { atomic.AddInt64(&GlobalMetrics.CircuitBreakerTrips, 1) }

func MetricsHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(GlobalMetrics)
}
