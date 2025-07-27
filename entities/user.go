package entities

import (
	"github.com/DanielPetcov/chatapp_go/auth"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// types
type UserHandler struct {
	JWTHandler *auth.JWTHandler
	UsersColl  *mongo.Collection
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		JWTHandler: auth.NewJWTHandler(),
	}
}
