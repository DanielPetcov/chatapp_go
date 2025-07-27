package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/DanielPetcov/chatapp_go/auth"
	"github.com/DanielPetcov/chatapp_go/entities"
	"github.com/DanielPetcov/chatapp_go/general"
	"github.com/DanielPetcov/chatapp_go/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
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
		AllowOrigins:     []string{"https://chat-app-go-frontend.vercel.app", "https://api.ppeettss.shop", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
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

	chatHandler := entities.NewChatHandler()
	chatHandler.ChatColl = db.Collection("chats")

	v1.GET("/chat", chatHandler.ListOfChats)
	v1.POST("/chat", chatHandler.NewChat)
	v1.POST("/chat/user", chatHandler.AddToChat)

	hub := websocket.NewHub()
	hub.ChatColl = db.Collection("chats")
	go hub.Run()

	jwtHandler := auth.NewJWTHandler()

	v1.GET("/ws", func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			general.GeneralError(ctx, err)
			return
		}

		userID, err := jwtHandler.CheckJWT(token)
		if err != nil {
			general.GeneralError(ctx, err)
			return
		}
		userObjectID, err := bson.ObjectIDFromHex(userID)
		if err != nil {
			general.GeneralError(ctx, err)
			return
		}

		websocket.ServeWs(hub, ctx, userObjectID)
	})
	router.Run(":8080")
}
