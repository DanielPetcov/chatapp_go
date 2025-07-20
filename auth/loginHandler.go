package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHanlder(c *gin.Context) {
	var data = LoginBody{}
	if err := c.BindJSON(&data); err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"message": "error",
		})
		return
	}

	fmt.Println(data.Username)
	fmt.Println(data.Password)

	c.JSON(200, gin.H{
		"message": "logged",
	})
}
