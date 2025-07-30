package entities

import (
	"context"
	"log"

	"github.com/DanielPetcov/chatapp_go/auth"
	"github.com/DanielPetcov/chatapp_go/general"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// types
type ChatHandler struct {
	ChatColl   *mongo.Collection
	JWTHandler *auth.JWTHandler
}

type Message struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Text        string        `bson:"text,omitemtpy"`
	DateCreated bson.DateTime `bson:"dateCreated,omitempty"`
	Author      bson.ObjectID `bson:"author,omitempty"`
}

type ChatDB struct {
	ID             bson.ObjectID   `bson:"_id,omitempty"`
	Name           string          `bson:"name,omitempty"`
	AuthorID       bson.ObjectID   `bson:"authorID,omitempty"`
	ParticipantsID []bson.ObjectID `bson:"participantsID,omitempty"`
	Messages       []Message       `bson:"messages,omitempty"`
}

type NewChatBody struct {
	Name string `json:"name"`
}

type RemoveUserFromChatBody struct {
	ChatID string `json:"chatID"`
	UserID string `json:"userID"`
}

type AddToChatBody struct {
	ChatID string `json:"chatID"`
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		JWTHandler: auth.NewJWTHandler(),
	}
}

type ChatMessagesBody struct {
	Messages []Message `json:"messages,omitempty"`
}

func (c *ChatHandler) ListOfChats(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	token, err := c.JWTHandler.ExtractJWTfromAuth(authHeader)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userID, err := c.JWTHandler.CheckJWT(token)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userObjectId, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	filter := bson.D{
		{
			Key:   "participantsID",
			Value: userObjectId,
		},
	}

	cursor, err := c.ChatColl.Find(context.Background(), filter)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	var chats = []ChatDB{}
	if err := cursor.All(context.Background(), &chats); err != nil {
		general.GeneralError(ctx, err)
		return
	}
	ctx.JSON(200, gin.H{
		"message": "ok",
		"chats":   chats,
	})
}

func (c *ChatHandler) NewChat(ctx *gin.Context) {
	var data NewChatBody
	if err := ctx.BindJSON(&data); err != nil {
		general.GeneralError(ctx, err)
		return
	}

	authHeader := ctx.GetHeader("Authorization")
	token, err := c.JWTHandler.ExtractJWTfromAuth(authHeader)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userId, err := c.JWTHandler.CheckJWT(token)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userObjectID, err := bson.ObjectIDFromHex(userId)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	_, err = c.ChatColl.InsertOne(context.Background(), ChatDB{
		Name:           data.Name,
		AuthorID:       userObjectID,
		ParticipantsID: []bson.ObjectID{userObjectID},
	})

	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{
		"message": "ok",
	})
}

func (c *ChatHandler) RemoveUserFromChat(ctx *gin.Context) {
	var dataBody RemoveUserFromChatBody
	err := ctx.BindJSON(&dataBody)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	authHeader := ctx.GetHeader("Authorization")

	token, err := c.JWTHandler.ExtractJWTfromAuth(authHeader)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	// if jwt bad, then exit
	_, err = c.JWTHandler.CheckJWT(token)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	chatObjectID, err := bson.ObjectIDFromHex(dataBody.ChatID)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userObjectID, err := bson.ObjectIDFromHex(dataBody.UserID)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	var chat ChatDB
	err = c.ChatColl.FindOne(context.Background(), bson.M{"_id": chatObjectID}).Decode(&chat)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	if chat.AuthorID == userObjectID {
		_, err := c.ChatColl.DeleteOne(context.Background(), bson.M{"_id": chatObjectID})
		if err != nil {
			general.GeneralError(ctx, err)
			return
		}
	} else {
		update := bson.D{{
			Key: "$pull",
			Value: bson.D{{
				Key:   "participantsID",
				Value: userObjectID,
			}},
		}}

		_, err := c.ChatColl.UpdateOne(context.Background(), bson.M{"_id": chatObjectID}, update)
		if err != nil {
			general.GeneralError(ctx, err)
			return
		}
	}

	ctx.JSON(200, gin.H{
		"message": "ok",
	})
}

func (c *ChatHandler) AddToChat(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	token, err := c.JWTHandler.ExtractJWTfromAuth(authHeader)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userID, err := c.JWTHandler.CheckJWT(token)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	var chatBody AddToChatBody
	ctx.BindJSON(&chatBody)

	chatObjectID, err := bson.ObjectIDFromHex(chatBody.ChatID)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	filter := bson.D{
		{
			Key:   "_id",
			Value: chatObjectID,
		},
	}

	update := bson.D{
		{
			Key: "$addToSet",
			Value: bson.D{
				{Key: "participantsID", Value: userObjectID},
			},
		},
	}

	_, err = c.ChatColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{
		"message": "ok",
	})
}

func (c *ChatHandler) GetMessagesChat(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		log.Println("missing error")
		ctx.JSON(400, gin.H{
			"message": "missing chat id",
		})
	}

	chatObjectId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	filter := bson.D{
		{
			Key:   "_id",
			Value: chatObjectId,
		},
	}
	cursor := c.ChatColl.FindOne(context.Background(), filter)

	var chat ChatMessagesBody
	err = cursor.Decode(&chat)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{
		"messages": chat.Messages,
	})
}
