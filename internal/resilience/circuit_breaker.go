ackage resilience

import (
	"errors"
	"sync"
	"time"
	"weather-gateway/internal/metrics"
)

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"
	StateOpen     CircuitState = "OPEN"
	StateHalfOpen CircuitState = "HALF_OPEN"
)

type CircuitBreaker struct {
	mutex             sync.Mutex
	state             CircuitState
	failureCount      int
	failureThreshold  int
	halfOpenMaxProbes int
	halfOpenSuccesses int
	cooldownDuration  time.Duration
	lastStateChange   time.Time
}

func NewCircuitBreaker(threshold int, maxProbes int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:             StateClosed,
		failureThreshold:  threshold,
		halfOpenMaxProbes: maxProbes,
		cooldownDuration:  cooldown,
		lastStateChange:   time.Now(),
	}
}

func (cb *CircuitBreaker) Execute(operation func() (any, error)) (any, error) {
	if !cb.allowExecution() {
		return nil, errors.New("CIRCUIT_BREAKER_OPEN: system degraded, requests blocked via circuit breaker")
	}

	result, err := operation()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.failureCount++
		if cb.state == StateHalfOpen || cb.failureCount >= cb.failureThreshold {
			if cb.state != StateOpen {
				metrics.IncrementCircuitBreakerTrips()
				cb.state = StateOpen
				cb.lastStateChange = time.Now()
			}
		}
		return nil, err
	}

	if cb.state == StateHalfOpen {
		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.halfOpenMaxProbes {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.halfOpenSuccesses = 0
		}
	}
	return result, nil
}

func (cb *CircuitBreaker) allowExecution() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) > cb.cooldownDuration {
			cb.state = StateHalfOpen
			cb.halfOpenSuccesses = 0
			cb.lastStateChange = time.Now()
			return true
		}
		return false
	}
	return true
}
