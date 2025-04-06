package main

import (
	"context"
	"log"
	"os"
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

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)

	go serveWebSocket(ctx, waitGroup)
	go serveHtml(ctx)

	<-ctx.Done()

	waitGroup.Wait()

	log.Println("exit")
}
