package service

import (
	"strings"

	telegramapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleWeatherCommand 處理 /weather [城市] 指令
func handleWeatherCommand(botClient *telegramapi.BotAPI, message *telegramapi.Message) {
	chatID := message.Chat.ID

	city := strings.TrimSpace(message.CommandArguments())
	if city == "" {
		city = getUserDefaultCity(chatID)
	}

	weatherData, _, err := GetCurrentWeather(normalizeCityName(city))
	if err != nil {
		botClient.Send(telegramapi.NewMessage(chatID, "❌ 取得天氣失敗，請稍後再試"))
		return
	}

	botClient.Send(telegramapi.NewMessage(chatID, formatWeatherMessage(weatherData)))
}

// handleSetCityCommand 處理 /setcity 城市名 指令
func handleSetCityCommand(botClient *telegramapi.BotAPI, message *telegramapi.Message) {
	chatID := message.Chat.ID

	city := strings.TrimSpace(message.CommandArguments())
	if city == "" {
		botClient.Send(telegramapi.NewMessage(chatID, "請輸入城市名稱，例如：/setcity 臺南市"))
		return
	}

	city = normalizeCityName(city)
	setUserDefaultCity(chatID, city)
	botClient.Send(telegramapi.NewMessage(chatID, "✅ 已設定預設城市為："+city))
}

// handleImageCommand 處理 /image 指令
func handleImageCommand(botClient *telegramapi.BotAPI, message *telegramapi.Message) {
	chatID := message.Chat.ID

	photo, _, err := GetRandomPhoto()
	if err != nil {
		botClient.Send(telegramapi.NewMessage(chatID, "❌ 取得圖片失敗，請稍後再試"))
		return
	}

	botClient.Send(telegramapi.NewMessage(chatID, photo.URLs.Regular))
}
