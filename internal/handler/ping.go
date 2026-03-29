package handler

import (
	"net/http"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Ping 處理 GET /ping 請求
// 回傳 pong 確認伺服器正常，並發送測試訊息到 Telegram
func Ping(c *gin.Context) {
	chatID, err := config.TelegramChatID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := service.SendMessage(chatID, "👋 測試訊息！yo0-backend 連線正常。"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pong", "telegram": "sent"})
}
