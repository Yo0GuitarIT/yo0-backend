package model

type WeatherPeriod struct {
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	Weather         string `json:"weather"`
	RainProbability string `json:"rainProbability"`
	MinTempC        string `json:"minTempC"`
	MaxTempC        string `json:"maxTempC"`
	Comfort         string `json:"comfort"`
}

type WeatherTimeRange struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Hours int    `json:"hours"`
}

type WeatherSummary struct {
	LocationName string           `json:"locationName"`
	TimeRange    WeatherTimeRange `json:"timeRange"`
	Periods      []WeatherPeriod  `json:"periods"`
	Current      *WeatherPeriod   `json:"current,omitempty"`
}
