package config

import (
	"fmt"
	"os"
	"strconv"
)

func TelegramBotToken() string {
	return os.Getenv("TELEGRAM_BOT_TOKEN")
}

func TelegramChatID() (int64, error) {
	s := os.Getenv("TELEGRAM_CHAT_ID")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("TELEGRAM_CHAT_ID 設定錯誤: %w", err)
	}
	return id, nil
}

func WeatherAPIKey() (string, error) {
	key := os.Getenv("CWB_API_KEY")
	if key == "" {
		key = os.Getenv("CWA_API_KEY")
	}
	if key == "" {
		return "", fmt.Errorf("CWB_API_KEY / CWA_API_KEY 未設定")
	}
	return key, nil
}

func UnsplashAccessKey() string {
	return os.Getenv("UNSPLASH_ACCESS_KEY")
}
