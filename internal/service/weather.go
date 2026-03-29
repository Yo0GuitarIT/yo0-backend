package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	"github.com/Yo0GuitarIT/yo0-backend/internal/model"
)

// GetCurrentWeather 呼叫 CWA 開放資料 API 取得指定城市天氣
// 回傳：清洗後的 24 小時重點資料、HTTP 狀態碼、錯誤
func GetCurrentWeather(locationName string) (*model.WeatherSummary, int, error) {
	apiKey, err := config.WeatherAPIKey()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	taipei, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return nil, 500, err
	}

	now := time.Now().In(taipei)
	timeTo := now.Add(24 * time.Hour)

	baseURL := "https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-C0032-001"
	query := url.Values{}
	query.Set("Authorization", apiKey)
	query.Set("timeFrom", now.Format("2006-01-02T15:04:05"))
	query.Set("timeTo", timeTo.Format("2006-01-02T15:04:05"))
	if locationName != "" {
		query.Set("locationName", locationName)
	}

	req, err := http.NewRequest("GET", baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, 500, err
	}
	req.Header.Set("accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("CWA API 回傳 %d", resp.StatusCode)
	}

	cleaned, err := extractWeatherSummary(raw, locationName, now, timeTo, taipei)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return cleaned, http.StatusOK, nil
}

func extractWeatherSummary(raw map[string]interface{}, locationName string, from, to time.Time, loc *time.Location) (*model.WeatherSummary, error) {
	records, ok := raw["records"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("CWA response 缺少 records")
	}

	locations, ok := records["location"].([]interface{})
	if !ok || len(locations) == 0 {
		return &model.WeatherSummary{
			LocationName: locationName,
			TimeRange: model.WeatherTimeRange{
				From:  from.Format(time.RFC3339),
				To:    to.Format(time.RFC3339),
				Hours: 24,
			},
			Periods: []model.WeatherPeriod{},
		}, nil
	}

	firstLocation, ok := locations[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("CWA location 格式錯誤")
	}

	resolvedLocationName, _ := firstLocation["locationName"].(string)
	if resolvedLocationName == "" {
		resolvedLocationName = locationName
	}

	weatherElements, ok := firstLocation["weatherElement"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("CWA weatherElement 格式錯誤")
	}

	type periodEntry struct {
		start time.Time
		end   time.Time
		data  model.WeatherPeriod
	}

	periodMap := map[string]*periodEntry{}

	for _, elementItem := range weatherElements {
		element, ok := elementItem.(map[string]interface{})
		if !ok {
			continue
		}

		elementName, _ := element["elementName"].(string)
		times, ok := element["time"].([]interface{})
		if !ok {
			continue
		}

		for _, timeItem := range times {
			timeEntry, ok := timeItem.(map[string]interface{})
			if !ok {
				continue
			}

			startStr, _ := timeEntry["startTime"].(string)
			endStr, _ := timeEntry["endTime"].(string)
			if startStr == "" || endStr == "" {
				continue
			}

			startAt, err := time.ParseInLocation("2006-01-02 15:04:05", startStr, loc)
			if err != nil {
				continue
			}
			endAt, err := time.ParseInLocation("2006-01-02 15:04:05", endStr, loc)
			if err != nil {
				continue
			}

			key := startStr + "|" + endStr
			entry, exists := periodMap[key]
			if !exists {
				entry = &periodEntry{
					start: startAt,
					end:   endAt,
					data: model.WeatherPeriod{
						StartTime: startAt.Format(time.RFC3339),
						EndTime:   endAt.Format(time.RFC3339),
					},
				}
				periodMap[key] = entry
			}

			parameter, _ := timeEntry["parameter"].(map[string]interface{})
			paramName, _ := parameter["parameterName"].(string)

			switch elementName {
			case "Wx":
				entry.data.Weather = paramName
			case "PoP":
				entry.data.RainProbability = paramName
			case "MinT":
				entry.data.MinTempC = paramName
			case "MaxT":
				entry.data.MaxTempC = paramName
			case "CI":
				entry.data.Comfort = paramName
			}
		}
	}

	entries := make([]*periodEntry, 0, len(periodMap))
	for _, e := range periodMap {
		if e.end.After(from) && e.start.Before(to) {
			entries = append(entries, e)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].start.Before(entries[j].start)
	})

	periods := make([]model.WeatherPeriod, 0, len(entries))
	for _, e := range entries {
		periods = append(periods, e.data)
	}

	summary := &model.WeatherSummary{
		LocationName: resolvedLocationName,
		TimeRange: model.WeatherTimeRange{
			From:  from.Format(time.RFC3339),
			To:    to.Format(time.RFC3339),
			Hours: 24,
		},
		Periods: periods,
	}

	if len(periods) > 0 {
		p := periods[0]
		summary.Current = &p
	}

	return summary, nil
}