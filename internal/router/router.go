package router

import (
	"github.com/Yo0GuitarIT/yo0-backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// Setup 註冊所有路由到 Gin 引擎
func Setup(routerEngine *gin.Engine) {
	// 健康檢查
	routerEngine.GET("/", func(context *gin.Context) {
		context.JSON(200, gin.H{"message": "yo0-backend is running!"})
	})

	// 測試用：確認伺服器對 Telegram 連線正常
	routerEngine.GET("/ping", handler.Ping)

	// /photos 路由群組
	photosGroup := routerEngine.Group("/photos")
	{
		photosGroup.GET("/random", handler.GetRandomPhoto) // GET /photos/random
	}

	// weather 路由群組
	weatherGroup := routerEngine.Group("/weather")
	{
		weatherGroup.GET("/current", handler.GetCurrentWeather) // GET /weather/current
	}

	// notify 路由群組
	notifyGroup := routerEngine.Group("/notify")
	{
		notifyGroup.POST("/test", handler.TestMorningPush) // POST /notify/test
	}
}
