package service

import (
	"log"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	telegramapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

// StartScheduler 啟動定時排程，每天台灣時間 06:00 發送早安推播
func StartScheduler(botClient *telegramapi.BotAPI) {
	_ = botClient

	chatID, err := config.TelegramChatID()
	if err != nil {
		log.Printf("[Scheduler] %v", err)
		return
	}

	cronScheduler := cron.New()

	// CRON_TZ=Asia/Taipei 確保不受伺服器時區影響
	// 0 6 * * * = 每天 06:00
	cronScheduler.AddFunc("CRON_TZ=Asia/Taipei 0 6 * * *", func() {
		if err := SendMorningPush(chatID); err != nil {
			log.Printf("[Scheduler] 推播失敗: %v", err)
			return
		}
		log.Printf("[Scheduler] 已完成早安推播")
	})

	cronScheduler.Start()
	log.Printf("[Scheduler] 定時排程已啟動，每天台灣時間 06:00 發送")
}
