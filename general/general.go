package general

import (
	"log"

	"github.com/gin-gonic/gin"
)

func GeneralError(ctx *gin.Context, err error) {
	log.Println(err)
	ctx.JSON(400, gin.H{
		"message": "error",
	})
}
