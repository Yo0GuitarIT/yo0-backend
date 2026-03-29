package service

import (
	"fmt"
	"log"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// botInstance 持有全域 bot 實例，供同 package 內的其他功能使用。
// 類似前端的 singleton pattern：let botClient: BotAPI | null = null
var botInstance *tgBotApi.BotAPI

// StartBot 啟動 Telegram Bot，使用 Long Polling 監聽訊息
func StartBot() error {
	bot, err := tgBotApi.NewBotAPI(config.TelegramBotToken())
	if err != nil {
		return fmt.Errorf("無法建立 bot: %w", err)
	}
	botInstance = bot

	log.Printf("Telegram bot 已啟動：@%s", bot.Self.UserName)

	// 啟動定時排程
	StartScheduler(bot)

	// 設定 Long Polling，timeout 60 秒
	u := tgBotApi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleCommand(bot, update.Message)
	}

	return nil
}

// handleCommand 根據指令名稱派發到對應的處理函式
// 類似前端 router 的概念：switch on command → handler
func handleCommand(bot *tgBotApi.BotAPI, msg *tgBotApi.Message) {
	chatID := msg.Chat.ID

	switch msg.Command() {
	case "menu":
		bot.Send(tgBotApi.NewMessage(chatID,
			"👋 yo0-backend bot 啟動成功！\n\n"+
				"可用指令：\n"+
				"/weather - 查詢預設城市 24 小時天氣\n"+
				"/weather 城市名 - 查指定城市（例：/weather 高雄市）\n"+
				"/setcity 城市名 - 設定你的預設城市\n"+
				"/mycity - 查看你的預設城市\n"+
				"/image - 取得一張隨機照片"))

	case "weather":
		handleWeatherCommand(bot, msg)

	case "setcity":
		handleSetCityCommand(bot, msg)

	case "mycity":
		city := getUserDefaultCity(chatID)
		bot.Send(tgBotApi.NewMessage(chatID, "你的預設城市是："+city))

	case "image":
		handleImageCommand(bot, msg)

	default:
		bot.Send(tgBotApi.NewMessage(chatID, "不支援的指令，請使用 /menu 查看可用功能"))
	}
}
