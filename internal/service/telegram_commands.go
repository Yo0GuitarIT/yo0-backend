package service

import (
	"fmt"
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

	botClient.Send(telegramapi.NewPhoto(chatID, telegramapi.FileURL(photo.URLs.Regular)))
}

// drawTarotPhoto 抽一張牌，組成可直接發送的 PhotoConfig（含正逆位處理）。
// 供 /tarot 指令與每日推播共用同一段邏輯。
func drawTarotPhoto(chatID int64) (telegramapi.PhotoConfig, error) {
	card, _, err := GetRandomTarot()
	if err != nil {
		return telegramapi.PhotoConfig{}, err
	}

	orientation := "正位"
	var photo telegramapi.PhotoConfig
	if card.Reversed {
		orientation = "逆位"
		// 逆位：下載牌面圖、旋轉 180° 後以 bytes 傳送
		imgBytes, rotErr := fetchRotatedImage(card.Image)
		if rotErr != nil {
			// 旋轉失敗就退回正向網址圖，至少不會發不出牌
			photo = telegramapi.NewPhoto(chatID, telegramapi.FileURL(card.Image))
		} else {
			photo = telegramapi.NewPhoto(chatID, telegramapi.FileBytes{Name: "tarot.jpg", Bytes: imgBytes})
		}
	} else {
		// 正位：直接給網址，由 Telegram 抓圖（不必下載）
		photo = telegramapi.NewPhoto(chatID, telegramapi.FileURL(card.Image))
	}

	photo.Caption = fmt.Sprintf("🔮 %s（%s）· %s\n%s", card.NameZh, card.Name, orientation, card.Meaning)
	return photo, nil
}

// handleTarotCommand 處理 /tarot 指令：抽一張牌並回傳牌面圖片
func handleTarotCommand(botClient *telegramapi.BotAPI, message *telegramapi.Message) {
	chatID := message.Chat.ID

	photo, err := drawTarotPhoto(chatID)
	if err != nil {
		botClient.Send(telegramapi.NewMessage(chatID, "❌ 抽牌失敗，請稍後再試"))
		return
	}

	botClient.Send(photo)
}
