package router

import (
	"github.com/Yo0GuitarIT/yo0-backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// Setup 註冊所有路由到 Gin 引擎
func Setup(r *gin.Engine) {
	// 健康檢查
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "yo0-backend is running!"})
	})

	// 測試用：確認伺服器對 Telegram 連線正常
	r.GET("/ping", handler.Ping)

	// /photos 路由群組
	photos := r.Group("/photos")
	{
		photos.GET("/random", handler.GetRandomPhoto) // GET /photos/random
	}

	// weather 路由群組
	weather := r.Group("/weather")
	{
		weather.GET("/current", handler.GetCurrentWeather) // GET /weather/current
	}

	// notify 路由群組
	notify := r.Group("/notify")
	{
		notify.POST("/test", handler.TestMorningPush) // POST /notify/test
	}
}
