# Weather Intelligence Gateway

A production-grade Go API Gateway built on top of the WeatherAI platform.

The gateway acts as a resilient middleware layer between client applications and the upstream WeatherAI API, providing caching, rate limiting, circuit breaking, request aggregation, observability, and operational safeguards required for high-throughput production systems.

---

## Project Overview

Most applications communicate directly with third-party APIs. While simple, this approach introduces several operational risks:

* Excessive upstream API calls
* Increased latency
* Cascading failures during provider outages
* Lack of observability
* Poor traffic control

This project introduces an intelligent gateway layer that protects both consumers and upstream services while improving performance and reliability.

---

## Core Features

### Weather Data Gateway

Retrieves current weather conditions and forecast data from the WeatherAI API.

### Request Aggregation

Combines multiple upstream requests into a single response to reduce client-side complexity.

### In-Memory TTL Cache

Frequently requested weather data is cached to reduce latency and minimize upstream API usage.

### Circuit Breaker

Protects the system from repeated upstream failures using CLOSED, OPEN, and HALF-OPEN states.

### Token Bucket Rate Limiter

Prevents abuse and controls traffic on a per-client basis.

### Structured Logging

Captures request metadata, latency, cache status, and error information using structured logs.

### Metrics Collection

Tracks:

* Total Requests
* Cache Hits
* Cache Misses
* Upstream Failures
* Rate-Limited Requests

### Health Monitoring

Provides health check endpoints suitable for Docker, Kubernetes, and load balancers.

### Graceful Shutdown

Ensures active requests complete safely during deployment or termination events.

---

## Architecture

Client Request

↓

Security Middleware

↓

Rate Limiter

↓

Singleflight Request Collapser

↓

TTL Cache

↓

Circuit Breaker

↓

WeatherAI API

---

## Repository Structure

```text
weather-gateway/

├── cmd/
│   └── gateway/
│       └── main.go

├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── gateway/
│   │   ├── handler.go
│   │   ├── middleware.go
│   │   └── routes.go
│   │
│   ├── resilience/
│   │   ├── circuit_breaker.go
│   │   └── rate_limiter.go
│   │
│   ├── weather/
│   │   ├── client.go
│   │   └── models.go
│   │
│   └── metrics/
│       └── metrics.go
│
├── pkg/
│   ├── cache/
│   │   └── ttl_cache.go
│   │
│   └── logger/
│       └── logger.go
│
├── Dockerfile
├── go.mod
├── README.md
└── .env.example
```

---

## WeatherAI Integration

Base URL:

```text
https://api.weather-ai.co
```

Authentication:

```http
Authorization: Bearer wai_<your_api_key>
```

Example upstream request:

```bash
curl \
-H "Authorization: Bearer wai_your_key" \
"https://api.weather-ai.co/v1/weather?lat=-1.2921&lon=36.8219"
```

---

## API Endpoints

### Weather Dashboard

Aggregates weather information and gateway metadata.

```http
GET /api/v1/weather/dashboard
```

Example:

```bash
curl "http://localhost:8080/api/v1/weather/dashboard?lat=-1.2921&lon=36.8219"
```

---

### Health Check

```http
GET /health
```

Response:

```json
{
  "status": "healthy"
}
```

---

### Metrics

```http
GET /metrics
```

Response:

```json
{
  "requests": 1523,
  "cache_hits": 1200,
  "cache_misses": 323,
  "upstream_failures": 7,
  "rate_limited": 15
}
```

---

## Configuration

Environment Variables

```bash
WEATHER_API_KEY=wai_your_api_key
SERVER_PORT=8080
CACHE_TTL=5m
RATE_LIMIT_PER_MINUTE=100
CIRCUIT_BREAKER_THRESHOLD=5
```

---

## Running Locally

Install dependencies:

```bash
go mod tidy
```

Export environment variables:

```bash
export WEATHER_API_KEY="wai_your_api_key"
export SERVER_PORT="8080"
```

Run the application:

```bash
go run cmd/gateway/main.go
```

Run with race detection:

```bash
go run -race cmd/gateway/main.go
```

---

## Testing

Run all tests:

```bash
go test ./...
```

Run verbose tests:

```bash
go test -v ./...
```

Run race-condition detection:

```bash
go test -race ./...
```

Run benchmarks:

```bash
go test -bench=. -benchmem ./...
```

---

## Docker Deployment

Build image:

```bash
docker build -t weather-gateway:latest .
```

Run container:

```bash
docker run -d \
-p 8080:8080 \
-e WEATHER_API_KEY="wai_your_api_key" \
-e SERVER_PORT="8080" \
weather-gateway:latest
```

---

## Scaling Considerations

### Horizontal Scaling

The service is stateless and can be scaled horizontally behind a load balancer.

### Distributed Caching

For multi-instance deployments, the in-memory cache can be replaced with Redis.

### Observability

Metrics can be exported to Prometheus and visualized using Grafana.

### Distributed Tracing

OpenTelemetry can be integrated to provide end-to-end request tracing.

### Circuit Protection

Circuit breaker patterns prevent cascading failures when upstream dependencies become unavailable.

---

## Design Decisions

### Why Cache?

Reduce latency and lower upstream API consumption.

### Why Rate Limiting?

Protect the platform from abuse and traffic spikes.

### Why Circuit Breakers?

Prevent repeated failures from overwhelming the gateway during upstream outages.

### Why Structured Logging?

Enable efficient debugging, monitoring, and production diagnostics.

---

## Future Enhancements

* Redis Distributed Cache
* Prometheus Metrics Exporter
* OpenTelemetry Tracing
* Kubernetes Helm Deployment
* Webhook Subscription Management
* WeatherAI Tree Analysis Integration
* Multi-Region Failover Support

---

## License

MIT License
