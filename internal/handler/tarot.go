package handler

import (
	"net/http"

	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GetRandomTarot 處理 GET /tarot/random 請求
func GetRandomTarot(context *gin.Context) {
	result, statusCode, err := service.GetRandomTarot()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(statusCode, result)
}