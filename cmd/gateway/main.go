package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather-gateway/internal/config"
	"weather-gateway/internal/gateway"
	"weather-gateway/internal/resilience"
	"weather-gateway/internal/weather"
	"weather-gateway/pkg/cache"
	"weather-gateway/pkg/logger"
)

func main() {
	logger.InitializeLogger()
	gatewayConfig := config.LoadConfig()

	logger.Info("Starting up API gateway engine core systems...", "port", gatewayConfig.ServerPort)

	// Central lifecycle context to cleanly stop asynchronous janitors
	applicationLifecycleCtx, cancelLifecycle := context.WithCancel(context.Background())
	defer cancelLifecycle()

	weatherClient := weather.NewUpstreamWeatherClient(gatewayConfig.WeatherAPIKey)
	circuitBreaker := resilience.NewCircuitBreaker(5, 3, 30*time.Second) // 5 drops opens, 3 trials in half-open state
	
	ttlCache := cache.NewTTLCache(gatewayConfig.MaxCacheSize)
	ttlCache.StartJanitor(applicationLifecycleCtx, 10*time.Minute)

	rateLimiter := resilience.NewIPTokenBucketLimiter(100.0, 100.0)
	rateLimiter.StartCleaner(applicationLifecycleCtx, 1*time.Hour)

	weatherHandler := gateway.NewWeatherGatewayHandler(weatherClient, circuitBreaker, ttlCache, gatewayConfig.CacheTTL)
	muxRoutes := gateway.ConfigureGatewayRoutes(weatherHandler, rateLimiter)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", gatewayConfig.ServerPort),
		Handler:      muxRoutes,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownSignalChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownSignalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Critical transport connection failure crashed the gateway server engine", err)
			os.Exit(1)
		}
	}()

	logger.Info("API Gateway successfully bound and running across proxy interfaces")

	<-shutdownSignalChannel
	logger.Info("Graceful termination signal received. Initializing system shutdown sequences...")

	// Halt background janitor routines instantly via context broad-casting loops
	cancelLifecycle()

	contextTimeoutGracePeriod, cancelContextTimeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelContextTimeout()

	if err := httpServer.Shutdown(contextTimeoutGracePeriod); err != nil {
		logger.Error("Forced shutdown sequence triggered due to active request process timeouts", err)
		os.Exit(1)
	}

	logger.Info("All operational request lines cleanly drained. API Gateway terminated gracefully.")
}
