package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/depuzhiguang/bot-service/internal/ai"
	"github.com/depuzhiguang/bot-service/internal/logger"
	"github.com/depuzhiguang/bot-service/internal/manager"
	"github.com/depuzhiguang/bot-service/internal/registrar"
	"github.com/depuzhiguang/bot-service/internal/scheduler"
)

func main() {
	logg := logger.New("bot-service")

	count := flag.Int("count", 20, "Number of sim users")
	wsURL := flag.String("ws", "ws://localhost:8080/ws", "Game server WebSocket URL")
	apiURL := flag.String("api", "http://localhost:3000", "API server base URL")
	dailyHands := flag.Int("hands", 1000, "Daily target hands")
	flag.Parse()

	if envURL := os.Getenv("GAME_SERVER_WS"); envURL != "" {
		*wsURL = envURL
	}
	if envAPI := os.Getenv("API_BASE_URL"); envAPI != "" {
		*apiURL = envAPI
	}

	logg.Info("simulation service starting")
	logg.Info("config", "users", *count, "ws", *wsURL, "api", *apiURL, "daily_hands", *dailyHands)

	// Step 1: Register users
	reg := registrar.NewRegistrar(*apiURL)
	profiles, _, err := reg.RegisterBatch(*count)
	if err != nil {
		logg.Error("failed to register users", "error", err)
		os.Exit(1)
	}
	logg.Info("registered users", "count", len(profiles))

	// Step 2: Setup scheduler
	userIDs := make([]string, len(profiles))
	for i, p := range profiles {
		userIDs[i] = p.UserID
	}
	sched, err := scheduler.NewScheduler(userIDs, 3, 5, 7)
	if err != nil {
		logg.Error("failed to create scheduler", "error", err)
		os.Exit(1)
	}

	// Step 3: Setup manager
	mgr := manager.NewManager(*wsURL, *apiURL)

	// Step 5: Create bots with personas
	personas := ai.AllPersonas()
	for i, profile := range profiles {
		style := personas[i%len(personas)]
		persona := ai.GetPersona(style)
		engine := ai.NewEngineWithPersona(persona, "BTN")
		// Store bot in manager
		mgr.RegisterBot(profile.UserID, engine)
	}

	// Step 6: Assign to tables and start
	tables := sched.Assign()
	for _, table := range tables {
		logg.Info("table assigned", "table_id", table.TableID, "user_count", len(table.Users))
		for _, uid := range table.Users {
			if err := mgr.AssignToTable(uid, table.TableID); err != nil {
				logg.Error("failed to assign user to table", "user_id", uid, "table_id", table.TableID, "error", err)
			}
		}
	}

	logg.Info("simulation running", "message", "Press Ctrl+C to stop")

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logg.Info("shutting down")
	mgr.StopAll()
	logg.Info("done")
}
