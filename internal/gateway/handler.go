package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"golang.org/x/sync/singleflight"
	"weather-gateway/internal/metrics"
	"weather-gateway/internal/resilience"
	"weather-gateway/internal/weather"
	"weather-gateway/pkg/cache"
	"weather-gateway/pkg/logger"
)

type WeatherGatewayHandler struct {
	weatherClient  weather.WeatherClient
	circuitBreaker *resilience.CircuitBreaker
	ttlCache       *cache.TTLCache
	cacheDuration  time.Duration
	requestGroup   singleflight.Group
}

func NewWeatherGatewayHandler(client weather.WeatherClient, cb *resilience.CircuitBreaker, appCache *cache.TTLCache, ttl time.Duration) *WeatherGatewayHandler {
	return &WeatherGatewayHandler{
		weatherClient:  client,
		circuitBreaker: cb,
		ttlCache:       appCache,
		cacheDuration:  ttl,
	}
}

func (handler *WeatherGatewayHandler) HandleGetWeatherDashboard(writer http.ResponseWriter, request *http.Request) {
	requestStartTime := time.Now()
	requestedCity := request.URL.Query().Get("city")
	requestID := writer.Header().Get("X-Request-ID")
	
	if requestedCity == "" {
		writer.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(writer).Encode(weather.APIErrorResponse{
			Code:      "BAD_REQUEST",
			Message:   "Missing required query string attribute: city",
			RequestID: requestID,
		})
		return
	}

	// 1. Look up cached assets
	if cachedWeather, hit := handler.ttlCache.Get(requestedCity); hit {
		metrics.IncrementCacheHits()
		logger.Info("Serving dashboard metrics directly from cache store", "city", requestedCity, "latency_ms", time.Since(requestStartTime).Milliseconds())
		json.NewEncoder(writer).Encode(cachedWeather)
		return
	}

	metrics.IncrementCacheMisses()

	// 2. High Scale Protection: singleflight collapses matching concurrent data stampedes
	sharedPayload, err, _ := handler.requestGroup.Do(requestedCity, func() (any, error) {
		// Enforce tight request execution timeout boundaries explicitly
		executionCtx, cancelExecution := context.WithTimeout(request.Context(), 3*time.Second)
		defer cancelExecution()

		return handler.circuitBreaker.Execute(func() (any, error) {
			return handler.weatherClient.FetchDashboardMetrics(executionCtx, requestedCity)
		})
	})

	if err != nil {
		metrics.IncrementUpstreamFailures()
		logger.Error("Upstream service orchestration boundary failure detected", err, "city", requestedCity, "request_id", requestID)
		writer.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(writer).Encode(weather.APIErrorResponse{
			Code:      "UPSTREAM_ERROR",
			Message:   fmt.Sprintf("Failed to resolve backend system integrations: %s", err.Error()),
			RequestID: requestID,
		})
		return
	}

	// 3. Update local feature cache metrics safely
	handler.ttlCache.Set(requestedCity, sharedPayload, handler.cacheDuration)

	logger.Info("Successfully unified and served aggregated engine metrics", 
		"city", requestedCity, 
		"latency_ms", time.Since(requestStartTime).Milliseconds(),
	)
	json.NewEncoder(writer).Encode(sharedPayload)
}

func HandleHealthProbe(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(`{"status":"healthy"}`))
}
