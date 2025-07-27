package auth

import (
	"context"

	"github.com/DanielPetcov/chatapp_go/general"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuthHandler struct {
	UsersColl  *mongo.Collection
	JWTHandler *JWTHandler
}

func NewAuthHandler() *AuthHandler {
	jwtHandler := NewJWTHandler()

	return &AuthHandler{
		JWTHandler: jwtHandler,
	}
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// database objects
type CreatedUserDB struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	Username       string        `bson:"username,omitempty"`
	PasswordHashed string        `bson:"password,omitempty"`
}

type LoginedUserDB struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Username string        `bson:"username,omitempty"`
	Password string        `bson:"password"`
}

type VerifyJWTBody struct {
	Token string `json:"token"`
}

func (a *AuthHandler) LoginHanlder(c *gin.Context) {
	var data = LoginBody{}
	if err := c.BindJSON(&data); err != nil {
		general.GeneralError(c, err)
		return
	}

	//database
	filter := bson.D{
		{
			Key:   "username",
			Value: data.Username,
		},
	}

	var userDB LoginedUserDB
	err := a.UsersColl.FindOne(context.Background(), filter).Decode(&userDB)
	if err != nil {
		general.GeneralError(c, err)
		return
	}

	if !Verify(userDB.Password, data.Password) {
		general.GeneralError(c, err)
		return
	}

	token, err := a.JWTHandler.NewJWT(userDB.ID.Hex())
	if err != nil {
		general.GeneralError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"jwt":     token,
		"userID":  userDB.ID.Hex(),
	})
}

func (a *AuthHandler) RegisterHandler(c *gin.Context) {
	var data = RegisterBody{}
	if err := c.BindJSON(&data); err != nil {
		general.GeneralError(c, err)
		return
	}

	// database implementation
	hashedPassword, err := Hash(data.Password)
	if err != nil {
		general.GeneralError(c, err)
		return
	}

	result, err := a.UsersColl.InsertOne(context.Background(), CreatedUserDB{
		Username:       data.Username,
		PasswordHashed: hashedPassword,
	})

	if err != nil {
		general.GeneralError(c, err)
		return
	}

	userID := result.InsertedID.(bson.ObjectID).Hex()

	token, err := a.JWTHandler.NewJWT(userID)
	if err != nil {
		general.GeneralError(c, err)
		return
	}

	// if create succesfully user, then return jwt
	c.JSON(200, gin.H{
		"message": "success",
		"jwt":     token,
		"userID":  userID,
	})
}

func (a *AuthHandler) VerifyJWT(c *gin.Context) {
	var data = VerifyJWTBody{}
	if err := c.BindJSON(&data); err != nil {
		general.GeneralError(c, err)
		return
	}

	_, err := a.JWTHandler.CheckJWT(data.Token)
	if err != nil {
		general.GeneralError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}
