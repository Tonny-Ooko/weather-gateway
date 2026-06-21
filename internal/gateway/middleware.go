package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"weather-gateway/internal/metrics"
	"weather-gateway/internal/resilience"
	"weather-gateway/pkg/logger"
)

func SecurityAndRateLimitingMiddleware(limiter *resilience.IPTokenBucketLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			metrics.IncrementTotalRequests()

			//  Inject strict enterprise headers across security footprints
			writer.Header().Set("X-Content-Type-Options", "nosniff")
			writer.Header().Set("X-Frame-Options", "DENY")
			writer.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none';")
			writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			writer.Header().Set("Content-Type", "application/json")

			// Structural tracking for Request IDs
			requestIDBytes := make([]byte, 16)
			rand.Read(requestIDBytes)
			requestID := hex.EncodeToString(requestIDBytes)
			writer.Header().Set("X-Request-ID", requestID)

			clientIPAddress := request.RemoteAddr
			if forwardedIP := request.Header.Get("X-Forwarded-For"); forwardedIP != "" {
				clientIPAddress = strings.Split(forwardedIP, ",")[0]
			}

			if !limiter.Allow(clientIPAddress) {
				metrics.IncrementRateLimitedRequests()
				logger.Info("Rate limit threshold breached by inbound user footprint", "client_ip", clientIPAddress, "request_id", requestID)
				writer.WriteHeader(http.StatusTooManyRequests)
				writer.Write([]byte(`{"code":"TOO_MANY_REQUESTS","message":"Rate capacity boundary hit. Please retry later."}`))
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}
