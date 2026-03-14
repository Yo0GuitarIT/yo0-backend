package service

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartBot 啟動 Telegram Bot，使用 Long Polling 監聽訊息
// 不需要 HTTPS / ngrok，適合本地開發
func StartBot() error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		return fmt.Errorf("無法建立 bot: %w", err)
	}

	log.Printf("Telegram bot 已啟動：@%s", bot.Self.UserName)

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
