package handler

import (
	"os"
	"strconv"

	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GetRandomPhoto 處理 GET /photos/random 請求
// 呼叫 service 層取得照片資料後，直接回傳 Unsplash 的原始狀態碼與 JSON
func GetRandomPhoto(c *gin.Context) {
	result, statusCode, err := service.GetRandomPhoto()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// Ping 處理 GET /ping 請求
// 回傳 pong 確認伺服器正常，並發送測試訊息到 Telegram
func Ping(c *gin.Context) {
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(500, gin.H{"error": "TELEGRAM_CHAT_ID 設定錯誤"})
		return
	}

	if err := service.SendMessage(chatID, "👋 測試訊息！yo0-backend 連線正常。"); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "pong", "telegram": "sent"})
}
