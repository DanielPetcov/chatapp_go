package auth

import (
	"log"

	"github.com/gin-gonic/gin"
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

	token, err := a.JWTHandler.NewJWT()
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

	token, err := a.JWTHandler.NewJWT()
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

func (a *AuthHandler) VerifyJWT(c *gin.Context) {
	var data = VerifyJWTBody{}
	if err := c.BindJSON(&data); err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"message": "error parsing token",
		})
	}

	err := a.JWTHandler.CheckJWT(data.Token)
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
