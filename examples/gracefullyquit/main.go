// Author: lipixun
// Created Time : 2019-01-26 16:47:07
//
// File Name: main.go
// Description:
//

package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bytespirit/appframe-go/gracefullyquit"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	quiter := gracefullyquit.NewGracefullQuiter(context.Background(), gracefullyquit.WithQuitHandlerFunc(func() {
		log.Println("Run quit handler and wait for worker thread completed")
		wg.Wait()
		log.Println("Worker thread completed")
	}))

	go func(ctx context.Context) {
		log.Println("Worker thread started")
		<-ctx.Done()
		log.Println("Quit signal received, sleep 2s")
		time.Sleep(time.Second * 2)
		log.Println("Worker thread exited")
		wg.Done()
	}(quiter.LiveContext())

	log.Println("Wait for exit")
	quiter.WaitUntilExit(0)
	log.Println("Exit")
}
