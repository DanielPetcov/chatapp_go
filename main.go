package main

import (
	"github.com/DanielPetcov/chatapp_go/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := router.Group("/v1")

	v1.POST("/login", auth.LoginHanlder)

	router.Run(":8080")
}
