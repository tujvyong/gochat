package controllers

import (
	"encoding/json"
	"fmt"
	"gochat/config"
	"gochat/utils"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type JSONError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func APIError(w http.ResponseWriter, errMessage string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jsonError, err := json.Marshal(JSONError{Error: errMessage, Code: code})
	if err != nil {
		log.Panic(err)
	}
	w.Write(jsonError)
}

func forCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", config.Config.Origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var apiValidPath = regexp.MustCompile("^/ws/([A-Za-z0-9]+)$")

func ServeWs(w http.ResponseWriter, r *http.Request, hub *Hub) {
	m := apiValidPath.FindStringSubmatch(r.URL.Path)
	if len(m) == 0 {
		APIError(w, "Not Found", http.StatusNotFound)
		return
	}
	channelName := mux.Vars(r)["channelName"]

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Header.Get("Origin") == config.Config.Origin {
				return true
			}
			return false
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panic(err)
	}
	user, err := ConnectUser(RedisDB, channelName)
	if err != nil {
		log.Panic(err)
	}

	client := &Client{
		Hub:     hub,
		User:    user,
		Conn:    conn,
		Send:    make(chan []byte),
		Channel: channelName,
	}
	client.Hub.Register <- client

	// サーバーとしての処理。チャンネルがHubになければ、チャンネルをサブスクライブする
	if utils.IsExist(hub.Channels, channelName) == false {
		hub.SubscribeChannel(channelName)
	}

	go client.MessagePump()
	go client.writePump()
	go client.readPump()

	// 同じサーバー内しか参照できない
	hub.SendUserList(channelName)
	client.GetMessages(channelName)
}

func StartWebServer() error {
	router := mux.NewRouter()
	hub := NewHub()
	go hub.Run()

	router.HandleFunc("/ws/{channelName}", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, hub)
	})
	router.Use(forCORS)

	fmt.Printf("\n\033[32mStart Web Server on port :%v\033[0m\n", config.Config.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), router)
}
