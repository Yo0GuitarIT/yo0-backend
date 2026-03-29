package service

import "sync"

const fallbackCity = "臺南市"

// userDefaultCity 以 Chat ID 為 key，儲存每位用戶設定的預設城市。
// sync.Map 是 Go 的 thread-safe map，對應前端的 Map<number, string>
var userDefaultCity sync.Map

func getUserDefaultCity(chatID int64) string {
	if v, ok := userDefaultCity.Load(chatID); ok {
		if city, ok := v.(string); ok && city != "" {
			return city
		}
	}
	return fallbackCity
}

func setUserDefaultCity(chatID int64, city string) {
	userDefaultCity.Store(chatID, city)
}
