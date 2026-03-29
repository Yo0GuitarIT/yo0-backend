package handler

import (
	"net/http"

	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GetCurrentWeather 處理 GET /weather/current 請求
// 可用 query string locationName 指定城市，預設為臺南市
func GetCurrentWeather(context *gin.Context) {
	locationName := context.DefaultQuery("locationName", "臺南市")

	result, statusCode, err := service.GetCurrentWeather(locationName)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(statusCode, result)
}