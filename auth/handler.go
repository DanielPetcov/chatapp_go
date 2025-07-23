package auth

import (
	"context"
	"log"

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

//

type VerifyJWTBody struct {
	Token string `json:"token"`
}

func (a *AuthHandler) LoginHanlder(c *gin.Context) {
	var data = LoginBody{}
	if err := c.BindJSON(&data); err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
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
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	if !Verify(userDB.Password, data.Password) {
		log.Println("invalid password")
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	token, err := a.JWTHandler.NewJWT(userDB.ID.Hex())
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"jwt":     token,
	})
}

func (a *AuthHandler) RegisterHandler(c *gin.Context) {
	var data = RegisterBody{}
	if err := c.BindJSON(&data); err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	// database implementation
	hashedPassword, err := Hash(data.Password)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	result, err := a.UsersColl.InsertOne(context.Background(), CreatedUserDB{
		Username:       data.Username,
		PasswordHashed: hashedPassword,
	})

	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	userID := result.InsertedID.(bson.ObjectID).Hex()

	token, err := a.JWTHandler.NewJWT(userID)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	// if create succesfully user, then return jwt
	c.JSON(200, gin.H{
		"message": "success",
		"jwt":     token,
	})
}

func (a *AuthHandler) VerifyJWT(c *gin.Context) {
	var data = VerifyJWTBody{}
	if err := c.BindJSON(&data); err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error parsing token",
		})
	}

	_, err := a.JWTHandler.CheckJWT(data.Token)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "invalid token",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}
