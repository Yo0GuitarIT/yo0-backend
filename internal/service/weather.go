package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"
)

// GetCurrentWeather 呼叫 CWA 開放資料 API 取得指定城市天氣
// 回傳：清洗後的 24 小時重點資料、HTTP 狀態碼、錯誤
func GetCurrentWeather(locationName string) (map[string]interface{}, int, error) {
	apiKey := os.Getenv("CWB_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CWA_API_KEY")
	}
	if apiKey == "" {
		return nil, 500, fmt.Errorf("CWB_API_KEY/CWA_API_KEY 未設定")
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

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 500, err
	}

	if resp.StatusCode != http.StatusOK {
		return result, resp.StatusCode, nil
	}

	cleaned, err := extractWeatherSummary(result, locationName, now, timeTo, taipei)
	if err != nil {
		return nil, 500, err
	}

	return cleaned, resp.StatusCode, nil
}

func extractWeatherSummary(raw map[string]interface{}, locationName string, from, to time.Time, loc *time.Location) (map[string]interface{}, error) {
	records, ok := raw["records"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("CWA response 缺少 records")
	}

	locations, ok := records["location"].([]interface{})
	if !ok || len(locations) == 0 {
		return map[string]interface{}{
			"locationName": locationName,
			"timeRange": map[string]interface{}{
				"from": from.Format(time.RFC3339),
				"to":   to.Format(time.RFC3339),
			},
			"periods": []interface{}{},
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

	type period struct {
		Start time.Time
		End   time.Time
		Data  map[string]interface{}
	}

	periodMap := map[string]*period{}

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
			item, exists := periodMap[key]
			if !exists {
				item = &period{
					Start: startAt,
					End:   endAt,
					Data: map[string]interface{}{
						"startTime": startAt.Format(time.RFC3339),
						"endTime":   endAt.Format(time.RFC3339),
					},
				}
				periodMap[key] = item
			}

			parameter, _ := timeEntry["parameter"].(map[string]interface{})
			parameterName, _ := parameter["parameterName"].(string)

			switch elementName {
			case "Wx":
				item.Data["weather"] = parameterName
			case "PoP":
				item.Data["rainProbability"] = parameterName
			case "MinT":
				item.Data["minTempC"] = parameterName
			case "MaxT":
				item.Data["maxTempC"] = parameterName
			case "CI":
				item.Data["comfort"] = parameterName
			}
		}
	}

	periods := make([]*period, 0, len(periodMap))
	for _, p := range periodMap {
		if p.End.After(from) && p.Start.Before(to) {
			periods = append(periods, p)
		}
	}

	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Start.Before(periods[j].Start)
	})

	cleanedPeriods := make([]map[string]interface{}, 0, len(periods))
	for _, p := range periods {
		cleanedPeriods = append(cleanedPeriods, p.Data)
	}

	response := map[string]interface{}{
		"locationName": resolvedLocationName,
		"timeRange": map[string]interface{}{
			"from":  from.Format(time.RFC3339),
			"to":    to.Format(time.RFC3339),
			"hours": 24,
		},
		"periods": cleanedPeriods,
	}

	if len(cleanedPeriods) > 0 {
		response["current"] = cleanedPeriods[0]
	}

	return response, nil
}