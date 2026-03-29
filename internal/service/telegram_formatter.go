package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/Yo0GuitarIT/yo0-backend/internal/model"
)

// normalizeCityName 統一城市名稱格式（台 → 臺）
func normalizeCityName(city string) string {
	city = strings.TrimSpace(city)
	city = strings.ReplaceAll(city, "台", "臺")
	return city
}

// formatWeatherMessage 將 WeatherSummary 轉為 Telegram 訊息文字
func formatWeatherMessage(data *model.WeatherSummary) string {
	locationName := data.LocationName
	if locationName == "" {
		locationName = fallbackCity
	}

	messageText := fmt.Sprintf("🌤 %s 24 小時天氣\n", locationName)

	from := data.TimeRange.From
	to := data.TimeRange.To
	if from != "" && to != "" {
		messageText += fmt.Sprintf("⏱ 時間\n%s\n%s\n\n",
			formatRFC3339ForDisplay(from),
			formatRFC3339ForDisplay(to),
		)
	}

	if data.Current != nil {
		currentPeriod := data.Current
		messageText += fmt.Sprintf("目前時段：%s\n", safeString(currentPeriod.Weather, "資料不足"))
		messageText += fmt.Sprintf("🌧 降雨機率：%s%%\n", safeString(currentPeriod.RainProbability, "-"))
		messageText += fmt.Sprintf("🌡 溫度：%s°C ~ %s°C\n", safeString(currentPeriod.MinTempC, "-"), safeString(currentPeriod.MaxTempC, "-"))
		messageText += fmt.Sprintf("🙂 體感：%s", safeString(currentPeriod.Comfort, "-"))
	} else {
		messageText += "目前時段：資料不足"
	}

	return messageText
}

func safeString(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func formatRFC3339ForDisplay(raw string) string {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}
	return t.Format("2006-01-02 15:04")
}
