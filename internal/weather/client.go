package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WeatherClient interface {
	FetchDashboardMetrics(ctx context.Context, city string) (*WeatherDashboardResponse, error)
}

type UpstreamWeatherClient struct {
	apiToken   string
	httpClient *http.Client
}

func NewUpstreamWeatherClient(token string) WeatherClient {
	// Hardened transport client featuring custom pooling and connection boundaries
	customTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	return &UpstreamWeatherClient{
		apiToken: token,
		httpClient: &http.Client{
			Transport: customTransport,
			Timeout:   5 * time.Second,
		},
	}
}

func (client *UpstreamWeatherClient) FetchDashboardMetrics(ctx context.Context, city string) (*WeatherDashboardResponse, error) {
	// Production Pattern: Robust, structured downstream client utilizing context constraints
	targetURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", client.apiToken, city)
	
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate network integration request: %w", err)
	}

	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("upstream platform transaction failure: %w", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream gateway responded with non-200 state: %d", httpResponse.StatusCode)
	}

	var dataPayload WeatherDashboardResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&dataPayload); err != nil {
		return nil, fmt.Errorf("failed parsing downstream JSON payloads: %w", err)
	}

	return &dataPayload, nil
}
