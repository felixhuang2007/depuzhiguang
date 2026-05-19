package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/depuzhiguang/game-server/internal/server"
)

func main() {
	fmt.Println("De Pu Zhi Guang - Game Server")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := server.NewServer(":" + port)
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}
