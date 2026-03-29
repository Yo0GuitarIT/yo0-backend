package handler

import (
	"net/http"

	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GetRandomPhoto 處理 GET /photos/random 請求
func GetRandomPhoto(context *gin.Context) {
	result, statusCode, err := service.GetRandomPhoto()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(statusCode, result)
}
