package controllers

import (
	"encoding/json"
	"gochat/app/models"
	"gochat/utils"
	"log"
	"time"

	"github.com/go-redis/redis"
)

const channel_history_max = 10

var RedisDB *redis.Client

func init() {
	RedisDB = redis.NewClient(&redis.Options{
		Addr:     "gochat_redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func ZaddLog(channelName string, chatLog *ChatMessage) {
	now := time.Now().UnixNano()
	logsKey := utils.StrJoin(channelName, "_log")

	data, err := json.Marshal(chatLog)
	if err != nil {
		log.Panic(err, ContainerID)
	}
	if err := RedisDB.ZAdd(logsKey, redis.Z{
		Score:  float64(now),
		Member: data,
	}).Err(); err != nil {
		log.Panic(err, ContainerID)
	}
}

func PublishToRedis(channelName string, msg interface{}) {
	switch msg.(type) {
	case []*models.User:
		var users = &ServerSend{
			Command: "NEW_USER",
			Msg:     msg.([]*models.User),
		}
		data, err := json.Marshal(users)
		if err != nil {
			log.Panic(err, ContainerID)
		}
		err = RedisDB.Publish(channelName, data).Err()
		if err != nil {
			log.Panic(err, ContainerID)
		}
	case *ChatMessage:
		var newMsg = &ServerSend{
			Command: "MESSAGE",
			Msg:     msg,
		}
		data, err := json.Marshal(newMsg)
		if err != nil {
			log.Panic(err, ContainerID)
		}
		err = RedisDB.Publish(channelName, data).Err()
		if err != nil {
			log.Panic(err, ContainerID)
		}
		ZaddLog(channelName, msg.(*ChatMessage))
	default:
		log.Fatalf("Can't send message to redis %v: %T", channelName, msg)
	}
}
