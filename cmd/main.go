package main

import (
	"log"

	"github.com/Yo0GuitarIT/yo0-backend/internal/router"
	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// 背景執行 Telegram Bot（Long Polling）
	go func() {
		if err := service.StartBot(); err != nil {
			log.Printf("Telegram bot 錯誤: %v", err)
		}
	}()

	// 建立 Gin 引擎，附帶 Logger 和 Recovery middleware
	routerEngine := gin.Default()

	// 只信任本機（適用於無反向代理的直接部署）
	// 若前面有 nginx / Cloud Run 等反向代理，請改為對應的 IP 段
	routerEngine.SetTrustedProxies(nil)

	// 註冊所有路由
	router.Setup(routerEngine)

	// 啟動 HTTP 伺服器，監聽 8080 port
	routerEngine.Run(":8080")
}