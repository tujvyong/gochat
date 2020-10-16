package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type ConfigList struct {
	Port    int
	LogFile string
	Origin  string
}

var Config ConfigList

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln(err)
		return
	}

	Config = ConfigList{
		Port:    cfg.Section("web").Key("port").MustInt(),
		LogFile: cfg.Section("log").Key("file").String(),
		Origin:  cfg.Section("web").Key("origin").String(),
	}
}
