package weather

type CurrentWeatherModel struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
}

type ForecastModel struct {
	Summary string `json:"summary"`
}

type AirQualityModel struct {
	AQI float64 `json:"aqi"`
}

type WeatherDashboardResponse struct {
	City       string               `json:"city"`
	Weather    *CurrentWeatherModel `json:"weather"`
	Forecast   *ForecastModel       `json:"forecast"`
	AirQuality *AirQualityModel     `json:"air_quality"`
}

type APIErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
