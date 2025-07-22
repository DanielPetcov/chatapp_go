package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/DanielPetcov/chatapp_go/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongodbUrl := os.Getenv("MONGODB_URL")
	client, err := mongo.Connect(options.Client().ApplyURI(mongodbUrl))
	if err != nil {
		log.Fatal("failed to connect to db")
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Error on disconting the client")
		}
	}()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://chat-app-go-frontend.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	db := client.Database("chatapp")

	authHandler := auth.NewAuthHandler()
	authHandler.UsersColl = db.Collection("users")

	v1 := router.Group("/v1")
	v1.POST("/login", authHandler.LoginHanlder)
	v1.POST("/register", authHandler.RegisterHandler)
	v1.POST("/token", authHandler.VerifyJWT)

	router.Run(":8080")
}
