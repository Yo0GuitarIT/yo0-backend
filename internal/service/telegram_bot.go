package service

import (
	"fmt"
	"log"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	telegramapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// botInstance 持有全域 bot 實例，供同 package 內的其他功能使用。
// 類似前端的 singleton pattern：let botClient: BotAPI | null = null
var botInstance *telegramapi.BotAPI

// StartBot 啟動 Telegram Bot，使用 Long Polling 監聽訊息
func StartBot() error {
	botClient, err := telegramapi.NewBotAPI(config.TelegramBotToken())
	if err != nil {
		return fmt.Errorf("無法建立 bot: %w", err)
	}
	botInstance = botClient

	log.Printf("Telegram bot 已啟動：@%s", botClient.Self.UserName)

	// 啟動定時排程
	StartScheduler(botClient)

	// 設定 Long Polling，timeout 60 秒
	updateConfig := telegramapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updatesChannel := botClient.GetUpdatesChan(updateConfig)

	for update := range updatesChannel {
		if update.Message == nil {
			continue
		}
		handleCommand(botClient, update.Message)
	}

	return nil
}

// handleCommand 根據指令名稱派發到對應的處理函式
// 類似前端 router 的概念：switch on command → handler
func handleCommand(botClient *telegramapi.BotAPI, message *telegramapi.Message) {
	chatID := message.Chat.ID

	switch message.Command() {
	case "menu":
		botClient.Send(telegramapi.NewMessage(chatID,
			"👋 yo0-backend bot 啟動成功！\n\n"+
				"可用指令：\n"+
				"/weather - 查詢預設城市 24 小時天氣\n"+
				"/weather 城市名 - 查指定城市（例：/weather 高雄市）\n"+
				"/setcity 城市名 - 設定你的預設城市\n"+
				"/mycity - 查看你的預設城市\n"+
				"/image - 取得一張隨機照片"))

	case "weather":
		handleWeatherCommand(botClient, message)

	case "setcity":
		handleSetCityCommand(botClient, message)

	case "mycity":
		city := getUserDefaultCity(chatID)
		botClient.Send(telegramapi.NewMessage(chatID, "你的預設城市是："+city))

	case "image":
		handleImageCommand(botClient, message)

	default:
		botClient.Send(telegramapi.NewMessage(chatID, "不支援的指令，請使用 /menu 查看可用功能"))
	}
}
