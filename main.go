package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logFile, err := os.OpenFile("tabletop.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	config := loadConfig()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)

	go serveWebSocket(ctx, config, waitGroup)
	go serveHtml(ctx, config)

	<-ctx.Done()

	waitGroup.Wait()

	log.Println("exit")
}

type Config struct {
	ServerHost    string `json:"serverHost"`
	ServerPort    string `json:"serverPort"`
	WebsocketPort string `json:"websocketPort"`
}

func loadConfig() Config {
	var config Config

	configBytes, err := os.ReadFile("." + string(filepath.Separator) + "config.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(configBytes, &config); err != nil {
		panic(err)
	}

	log.Println("config", config)

	return config
}
