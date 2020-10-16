package main

import (
	"gochat/app/controllers"
	"gochat/config"
	"gochat/utils"
	"log"
	"net/http"
)

func fooHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("foo"))
}

func main() {
	utils.LogginSettings(config.Config.LogFile)

	log.Fatalln(controllers.StartWebServer())
}
