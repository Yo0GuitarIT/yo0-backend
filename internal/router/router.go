package router

import (
	"github.com/Yo0GuitarIT/yo0-backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "yo0-backend is running!"})
	})

	photos := r.Group("/photos")
	{
		photos.GET("/random", handler.GetRandomPhoto)
	}
}
