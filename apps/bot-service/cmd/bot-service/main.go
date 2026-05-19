package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/depuzhiguang/bot-service/internal/manager"
)

func main() {
	count := flag.Int("count", 100, "Number of bots")
	wsURL := flag.String("ws", "ws://localhost:8080/ws", "Game server WebSocket URL")
	flag.Parse()

	fmt.Printf("Bot Service starting with %d bots\n", *count)
	m := manager.NewManager(*wsURL)
	m.Spawn(*count)
	fmt.Printf("Spawned %d bots\n", *count)

	// Assign bots to default table
	assigned := 0
	for _, id := range m.BotIDs() {
		if err := m.AssignToTable(id, "default-6max"); err != nil {
			fmt.Printf("Failed to assign bot %s: %v\n", id, err)
		} else {
			assigned++
			if assigned >= 6 {
				break // fill one 6-max table for testing
			}
		}
	}
	fmt.Printf("Assigned %d bots to table\n", assigned)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
	m.StopAll()
	fmt.Println("All bots stopped")
}
