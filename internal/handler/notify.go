package handler

import (
	"net/http"
	"strconv"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// TestMorningPush 處理 POST /notify/test 請求
// 可用 query chatId 指定接收者，未提供則使用 TELEGRAM_CHAT_ID
func TestMorningPush(context *gin.Context) {
	chatIDQuery := context.Query("chatId")

	var (
		chatID int64
		err    error
	)

	if chatIDQuery == "" {
		chatID, err = config.TelegramChatID()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		chatID, err = strconv.ParseInt(chatIDQuery, 10, 64)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "chatId 格式錯誤，請提供數字"})
			return
		}
	}

	if err := service.SendMorningPush(chatID); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "測試推播已送出",
		"chatId":  chatID,
	})
}