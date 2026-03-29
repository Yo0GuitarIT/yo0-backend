package service

import (
	"fmt"

	telegramapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMessage 發送文字訊息到指定 Chat ID
func SendMessage(chatID int64, text string) error {
	if botInstance == nil {
		return fmt.Errorf("bot 尚未初始化")
	}
	_, err := botInstance.Send(telegramapi.NewMessage(chatID, text))
	return err
}

// SendMorningPush 發送早安推播（隨機照片 + 24 小時天氣）
// 供排程器與測試 API 共用同一段邏輯
func SendMorningPush(chatID int64) error {
	if botInstance == nil {
		return fmt.Errorf("bot 尚未初始化")
	}

	photo, _, err := GetRandomPhoto()
	if err != nil {
		return fmt.Errorf("取得照片失敗: %w", err)
	}

	imageURL := photo.URLs.Regular
	if _, err := botInstance.Send(telegramapi.NewMessage(chatID, "🌅 早安！今日隨機照片：\n"+imageURL)); err != nil {
		return fmt.Errorf("發送照片失敗: %w", err)
	}

	city := getUserDefaultCity(chatID)
	weatherData, _, err := GetCurrentWeather(normalizeCityName(city))
	if err != nil {
		return fmt.Errorf("取得天氣失敗: %w", err)
	}

	if _, err := botInstance.Send(telegramapi.NewMessage(chatID, "☀️ 早安！今日天氣預報\n\n"+formatWeatherMessage(weatherData))); err != nil {
		return fmt.Errorf("發送天氣失敗: %w", err)
	}

	return nil
}
