package main

import (
	"github.com/Yo0GuitarIT/yo0-backend/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	router.Setup(r)
	r.Run(":8080")
}