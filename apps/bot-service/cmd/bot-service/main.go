package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/depuzhiguang/bot-service/internal/manager"
)

func main() {
	count := flag.Int("count", 100, "Number of bots")
	flag.Parse()

	fmt.Printf("Bot Service starting with %d bots\n", *count)
	m := manager.NewManager()
	m.Spawn(*count)
	fmt.Printf("Spawned %d bots\n", *count)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
	m.StopAll()
	fmt.Println("All bots stopped")
}
