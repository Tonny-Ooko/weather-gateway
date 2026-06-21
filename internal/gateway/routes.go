package gateway

import (
	"net/http"
	"weather-gateway/internal/metrics"
	"weather-gateway/internal/resilience"
)

func ConfigureGatewayRoutes(handler *WeatherGatewayHandler, limiter *resilience.IPTokenBucketLimiter) *http.ServeMux {
	muxRouter := http.NewServeMux()

	rateLimitedChain := SecurityAndRateLimitingMiddleware(limiter)
	muxRouter.Handle("/api/v1/weather/dashboard", rateLimitedChain(http.HandlerFunc(handler.HandleGetWeatherDashboard)))

	muxRouter.HandleFunc("/health", HandleHealthProbe)
	muxRouter.HandleFunc("/metrics", metrics.MetricsHandler)

	return muxRouter
} 
