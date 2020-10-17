package controllers

import (
	"encoding/json"
	"fmt"
	"gochat/app/models"
	"log"

	"github.com/go-redis/redis"
)

// https://github.com/gorilla/websocket/tree/master/examples/chat
// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	Clients    map[string]*Client
	Channels   []*redis.PubSub
	Broadcast  chan *ClientSend
	Register   chan *Client
	Unregister chan *Client
	Received   chan *redis.Message
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[string]*Client),
		// Channels was allocated memory when "append" function was called
		Broadcast:  make(chan *ClientSend),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Received:   make(chan *redis.Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.User.Id] = client
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.User.Id]; ok {
				delete(h.Clients, client.User.Id)
				close(client.Send)
			}
		case userSend := <-h.Broadcast:
			var newMsg = &ChatMessage{
				Username: userSend.Username,
				Text:     userSend.Msg,
			}
			PublishToRedis(userSend.Channel, newMsg)
		case message := <-h.Received:
			for _, client := range h.Clients {
				if client.Channel == message.Channel {
					client.Send <- []byte(message.Payload)
				}
			}
		}
	}
}

func (h *Hub) SubscribeChannel(channelName string) {
	pubsub := RedisDB.Subscribe(channelName)
	h.Channels = append(h.Channels, pubsub)
}

func (h *Hub) GetPubsub(channelName string) *redis.PubSub {
	channelName = fmt.Sprintf("PubSub(%v)", channelName)
	for _, v := range h.Channels {
		if v.String() == channelName {
			return v
		}
	}
	return nil
}

func (h *Hub) GetUsers(channelName string) ([]*models.User, error) {
	usersData, err := RedisDB.HGetAll(channelName).Result()
	if err != nil {
		return nil, err
	}

	var users = make([]*models.User, len(usersData))
	i := 0
	for _, user := range usersData {
		var tmp *models.User
		if err := json.Unmarshal([]byte(user), &tmp); err != nil {
			log.Panic(err)
		}
		users[i] = tmp
		i++
	}
	return users, nil
}

func (h *Hub) SendUserList(channelName string) {
	users, err := h.GetUsers(channelName)
	if err != nil {
		log.Panic(err)
	}
	PublishToRedis(channelName, users)
}
