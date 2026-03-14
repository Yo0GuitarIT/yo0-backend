package service

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

// botInstance 持有全域 bot 實例，供其他 service 使用
var botInstance *tgbotapi.BotAPI

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

		switch update.Message.Command() {
		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"👋 yo0-backend bot 啟動成功！\n輸入 /image 取得一張隨機照片")
			bot.Send(msg)

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
		}
	}

	return nil
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
