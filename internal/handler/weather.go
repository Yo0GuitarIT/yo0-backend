package handler

import (
	"net/http"

	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// GetCurrentWeather 處理 GET /weather/current 請求
// 可用 query string locationName 指定城市，預設為臺南市
func GetCurrentWeather(c *gin.Context) {
	locationName := c.DefaultQuery("locationName", "臺南市")

	result, statusCode, err := service.GetCurrentWeather(locationName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}