package handler

import (
	"github.com/Yo0GuitarIT/yo0-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func GetRandomPhoto(c *gin.Context) {
	result, statusCode, err := service.GetRandomPhoto()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
