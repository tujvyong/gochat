package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochat/config"
	"gochat/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var ContainerID string

func init() {
	out, err := exec.Command("cat", "/etc/hostname").Output()
	if err != nil {
		log.Fatalln(err)
	}
	out = out[:len(out)-1]
	buf := bytes.NewBufferString("Container ID: ")
	buf.Write(out)
	ContainerID = buf.String()
}

type JSONError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func APIError(w http.ResponseWriter, errMessage string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jsonError, err := json.Marshal(JSONError{Error: errMessage, Code: code})
	if err != nil {
		log.Panic(err, ContainerID)
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
		log.Panic(err, ContainerID)
	}
	user, err := ConnectUser(RedisDB, channelName)
	if err != nil {
		log.Panic(err, ContainerID)
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

	hub.SendUserList(channelName)
	client.GetMessages(channelName)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	client.DisconnectUser(RedisDB)
	client.Conn.Close()
	os.Exit(1)
}

func StartWebServer() error {
	router := mux.NewRouter()
	hub := NewHub()
	go hub.Run()

	router.HandleFunc("/ws/{channelName}", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, hub)
	})
	router.Use(forCORS)

	log.Printf("Start Web Server on port :%v %v\n", config.Config.Port, ContainerID)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), router)
}
