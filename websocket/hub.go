package websocket

import (
	"context"
	"log"
	"time"

	"github.com/DanielPetcov/chatapp_go/entities"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Hub struct {
	clients map[*Client]bool

	broadcast chan SendContent

	register chan *Client

	unregister chan *Client

	ChatColl *mongo.Collection
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan SendContent),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			chatID, err := bson.ObjectIDFromHex(message.ChatID)
			if err != nil {
				log.Println(err)
				continue
			}
			filter := bson.M{"_id": chatID}

			cursor := h.ChatColl.FindOne(context.Background(), filter)
			if cursor == nil {
				log.Println("error finding a chat:")
				continue
			}

			var chat entities.ChatDB
			if err := cursor.Decode(&chat); err != nil {
				log.Printf("Error decoding chat: %v", err)
				continue
			}

			authorObjectID, err := bson.ObjectIDFromHex(message.UserID)
			if err != nil {
				log.Printf("Error decoding chat: %v", err)
				continue
			}

			h.ChatColl.UpdateByID(context.Background(), chat.ID, bson.M{
				"$push": bson.M{
					"messages": entities.Message{
						Text:        message.Message,
						DateCreated: bson.NewDateTimeFromTime(time.Now()),
						Author:      authorObjectID,
					},
				},
			})

			for client := range h.clients {
				for _, participantID := range chat.ParticipantsID {
					if participantID == client.ID {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client)
						}
						break
					}
				}
			}
		}
	}
}
