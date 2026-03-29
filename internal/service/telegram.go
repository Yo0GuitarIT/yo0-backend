package service

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

// botInstance 持有全域 bot 實例，供其他 service 使用
var botInstance *tgbotapi.BotAPI

var userDefaultCity sync.Map

const fallbackCity = "臺南市"

// SendMessage 發送訊息到指定 Chat ID
func SendMessage(chatID int64, text string) error {
	if botInstance == nil {
		return fmt.Errorf("bot 尚未初始化")
	}
	_, err := botInstance.Send(tgbotapi.NewMessage(chatID, text))
	return err
}

// StartBot 啟動 Telegram Bot，使用 Long Polling 監聽訊息
// 不需要 HTTPS / ngrok，適合本地開發
func StartBot() error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		return fmt.Errorf("無法建立 bot: %w", err)
	}
	botInstance = bot

	log.Printf("Telegram bot 已啟動：@%s", bot.Self.UserName)

	// 啟動定時排程（每天台灣時間 06:00 發送照片）
	StartScheduler(bot)

	// 設定 Long Polling，timeout 60 秒
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// 忽略非訊息的更新
		if update.Message == nil {
			continue
		}

		// 處理指令
		switch update.Message.Command() {
		case "menu":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"👋 yo0-backend bot 啟動成功！\n\n"+
					"可用指令：\n"+
					"/weather - 查詢預設城市 24 小時天氣\n"+
					"/weather 城市名 - 查指定城市（例：/weather 高雄市）\n"+
					"/setcity 城市名 - 設定你的預設城市\n"+
					"/mycity - 查看你的預設城市\n"+
					"/image - 取得一張隨機照片")
			bot.Send(msg)

		case "weather":
			args := strings.TrimSpace(update.Message.CommandArguments())
			city := args
			if city == "" {
				city = getUserDefaultCity(update.Message.Chat.ID)
			}

			weatherData, statusCode, err := GetCurrentWeather(normalizeCityName(city))
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ 取得天氣失敗，請稍後再試"))
				continue
			}

			if statusCode != 200 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ 天氣服務暫時不可用，請稍後再試"))
				continue
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, formatWeatherMessage(weatherData)))

		case "setcity":
			city := strings.TrimSpace(update.Message.CommandArguments())
			if city == "" {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "請輸入城市名稱，例如：/setcity 臺南市"))
				continue
			}

			city = normalizeCityName(city)
			userDefaultCity.Store(update.Message.Chat.ID, city)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "✅ 已設定預設城市為："+city))

		case "mycity":
			city := getUserDefaultCity(update.Message.Chat.ID)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "你的預設城市是："+city))

		case "image":
			// 呼叫現有的 unsplash service
			photo, _, err := GetRandomPhoto()
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ 取得圖片失敗，請稍後再試"))
				continue
			}

			urls, ok := photo["urls"].(map[string]interface{})
			if !ok {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ 圖片資料格式錯誤"))
				continue
			}

			imageURL, _ := urls["regular"].(string)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, imageURL))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"不支援的指令，請使用 /menu 查看可用功能"))
		}
	}

	return nil
}

func getUserDefaultCity(chatID int64) string {
	if v, ok := userDefaultCity.Load(chatID); ok {
		if city, ok := v.(string); ok && city != "" {
			return city
		}
	}
	return fallbackCity
}

func normalizeCityName(city string) string {
	city = strings.TrimSpace(city)
	city = strings.ReplaceAll(city, "台", "臺")
	return city
}

func formatWeatherMessage(data map[string]interface{}) string {
	locationName, _ := data["locationName"].(string)
	if locationName == "" {
		locationName = fallbackCity
	}

	timeRange, _ := data["timeRange"].(map[string]interface{})
	from, _ := timeRange["from"].(string)
	to, _ := timeRange["to"].(string)

	current, _ := data["current"].(map[string]interface{})
	weather, _ := current["weather"].(string)
	rainProbability, _ := current["rainProbability"].(string)
	minTemp, _ := current["minTempC"].(string)
	maxTemp, _ := current["maxTempC"].(string)
	comfort, _ := current["comfort"].(string)

	msg := fmt.Sprintf("🌤 %s 24 小時天氣\n", locationName)
	if from != "" && to != "" {
		msg += fmt.Sprintf("⏱ 時間\n%s\n%s\n\n", formatRFC3339ForDisplay(from), formatRFC3339ForDisplay(to))
	}

	msg += fmt.Sprintf("目前時段：%s\n", safeString(weather, "資料不足"))
	msg += fmt.Sprintf("🌧 降雨機率：%s%%\n", safeString(rainProbability, "-"))
	msg += fmt.Sprintf("🌡 溫度：%s°C ~ %s°C\n", safeString(minTemp, "-"), safeString(maxTemp, "-"))
	msg += fmt.Sprintf("🙂 體感：%s", safeString(comfort, "-"))

	return msg
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

// StartScheduler 啟動定時排程，每天台灣時間 06:00 自動發送隨機照片
// 需要環境變數 TELEGRAM_CHAT_ID 指定發送目標
func StartScheduler(bot *tgbotapi.BotAPI) {
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Printf("[Scheduler] TELEGRAM_CHAT_ID 設定錯誤: %v", err)
		return
	}

	c := cron.New()

	// CRON_TZ=Asia/Taipei 確保不受伺服器時區影響
	// 0 6 * * * = 每天 06:00
	c.AddFunc("CRON_TZ=Asia/Taipei 0 6 * * *", func() {
		photo, _, err := GetRandomPhoto()
		if err != nil {
			log.Printf("[Scheduler] 取得照片失敗: %v", err)
			return
		}

		urls, ok := photo["urls"].(map[string]interface{})
		if !ok {
			log.Printf("[Scheduler] 照片資料格式錯誤")
			return
		}

		imageURL, _ := urls["regular"].(string)
		bot.Send(tgbotapi.NewMessage(chatID, "🌅 早安！今日隨機照片：\n"+imageURL))
		log.Printf("[Scheduler] 已發送早安照片")
	})

	c.Start()
	log.Printf("[Scheduler] 定時排程已啟動，每天台灣時間 06:00 發送")
}
