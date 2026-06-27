package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

func DiscordBotToken() string {
	return os.Getenv("DISCORD_BOT_TOKEN")
}

// DiscordChannelIDs 讀取 DISCORD_CHANNEL_IDS（逗號分隔多個頻道 ID）。
// 若未設定則退而讀取舊的 DISCORD_CHANNEL_ID，以維持向下相容。
func DiscordChannelIDs() []string {
	s := os.Getenv("DISCORD_CHANNEL_IDS")
	if s == "" {
		if single := os.Getenv("DISCORD_CHANNEL_ID"); single != "" {
			return []string{single}
		}
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			ids = append(ids, trimmed)
		}
	}
	return ids
}
