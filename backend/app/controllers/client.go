package controllers

import (
	"encoding/json"
	"gochat/app/models"
	"gochat/utils"
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/icrowley/fake"
)

const (
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub     *Hub
	User    *models.User
	Conn    *websocket.Conn
	Send    chan []byte
	Channel string
}

type ServerSend struct {
	Command string      `json:"command"`
	Msg     interface{} `json:"msg"`
}

type ClientSend struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Msg      string `json:"msg"`
}

type ChatMessage struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		log.Println("readPump closed", c.User.Id, ContainerID)
		c.DisconnectUser(RedisDB)
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var userSend ClientSend
		if err := json.Unmarshal(message, &userSend); err != nil {
			log.Panic(err, ContainerID)
		}
		// Add username in here, because this program used fake user.
		userSend.Username = c.User.Username
		c.Hub.Broadcast <- &userSend
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		log.Println("writePump closed", c.User.Id, ContainerID)
		c.Hub.SendUserList(c.Channel)
		c.DisconnectUser(RedisDB)
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println(err)
				return
			}
			// log.Println(string(message))
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Println(err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (c *Client) MessagePump() {
	pubsub := c.Hub.GetPubsub(c.Channel)
	if pubsub == nil {
		log.Fatalf("Can't find pubsub channel on Hub: %v %v", c.Channel, ContainerID)
	}
	defer func() {
		pubsub.Close()
		c.DisconnectUser(RedisDB)
	}()

	for {
		iface, err := pubsub.Receive()
		if err != nil {
			log.Panic(err, ContainerID)
		}

		switch v := iface.(type) {
		case *redis.Subscription:
			if v.Kind == "unsubscribe" {
				log.Printf("Unsubscribe %v\n", v.Channel)
				break
			}
		case *redis.Message:
			c.Hub.Received <- v
		default:
			log.Printf("action=MessagePump: Recived unknown type \"%T\"", v)
		}
	}
}

func (c *Client) GetMessages(channelName string) {
	logsKey := utils.StrJoin(channelName, "_log")
	logs, err := RedisDB.ZRange(logsKey, -1*channel_history_max, -1).Result()
	if err != nil {
		log.Println(err)
		return
	}
	messages := make([]*ChatMessage, len(logs))
	for i := 0; i < len(logs); i++ {
		var tmp ChatMessage
		if err := json.Unmarshal([]byte(logs[i]), &tmp); err != nil {
			log.Panic(err, ContainerID)
		}
		messages[i] = &tmp
	}

	var messagesLog = &ServerSend{
		Command: "CHANNEL_LOG",
		Msg:     messages,
	}
	data, err := json.Marshal(messagesLog)
	if err != nil {
		log.Panic(err, ContainerID)
	}
	c.Send <- data
}

func ConnectUser(rdb *redis.Client, channelName string) (*models.User, error) {
	userId := models.RandString(16)

	userData := map[string]string{
		"id":       userId,
		"username": fake.UserName(),
	}
	data, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}
	if _, err := rdb.HSet(channelName, userId, string(data)).Result(); err != nil {
		return nil, err
	}

	var u *models.User
	err = json.Unmarshal(data, &u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (c *Client) DisconnectUser(rdb *redis.Client) {
	if err := rdb.HDel(c.Channel, c.User.Id).Err(); err != nil {
		log.Panic(err)
	}
}
