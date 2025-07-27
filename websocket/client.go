package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DanielPetcov/chatapp_go/general"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline = "\n"
	space   = " "
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SendContent struct {
	Message string `json:"message"`
	ChatID  string `json:"chatID"`
	UserID  string `json:"userID"`
}

type ReceivingContent struct {
	Message string `json:"message"`
	ChatID  string `json:"chatID"`
	UserID  string `json:"userID"`
}

type Client struct {
	ID bson.ObjectID

	hub *Hub

	conn *websocket.Conn

	send chan SendContent
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var content ReceivingContent
		if err := json.Unmarshal(rawMessage, &content); err != nil {
			log.Printf("unmarshall error: %v", err)
			continue
		}

		content.Message = strings.TrimSpace(strings.Replace(content.Message, newline, space, -1))
		c.hub.broadcast <- SendContent{
			Message: content.Message,
			ChatID:  content.ChatID,
			UserID:  content.UserID,
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case content, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsonContent, err := json.Marshal(content)
			if err != nil {
				log.Printf("marshall error: %v", err)
				continue
			}

			w.Write(jsonContent)

			n := len(c.send)
			for i := 0; i < n; i++ {
				localContent, ok := <-c.send
				if !ok {
					c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				jsonContentLocal, err := json.Marshal(localContent)
				if err != nil {
					log.Printf("marshall error: %v", err)
					break
				}

				w.Write(jsonContentLocal)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}

	}
}

func ServeWs(hub *Hub, ctx *gin.Context, userObjectID bson.ObjectID) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		general.GeneralError(ctx, err)
		return
	}

	client := &Client{
		ID:   userObjectID,
		hub:  hub,
		conn: conn,
		send: make(chan SendContent, 256),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
