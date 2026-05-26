package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	profiles, tokens, err := reg.RegisterBatch(*count)
	if err != nil {
		logg.Error("failed to register users", "error", err)
		os.Exit(1)
	}
	logg.Info("registered users", "count", len(profiles))

	// Step 2: Setup dynamic scheduler
	userIDs := make([]string, len(profiles))
	tokenMap := make(map[string]string, len(profiles))
	for i, p := range profiles {
		userIDs[i] = p.UserID
		tokenMap[p.UserID] = tokens[i]
	}
	ds := scheduler.NewDynamicScheduler(*apiURL, userIDs)

	// Step 3: Setup manager
	mgr := manager.NewManager(*wsURL, *apiURL, tokenMap)

	// Step 5: Create bots with personas
	personas := ai.AllPersonas()
	for i, profile := range profiles {
		style := personas[i%len(personas)]
		persona := ai.GetPersona(style)
		engine := ai.NewEngineWithPersona(persona, "BTN")
		// Store bot in manager
		mgr.RegisterBot(profile.UserID, engine)
	}

	logg.Info("simulation running", "message", "Press Ctrl+C to stop")

	// Step 6: Start dynamic scheduler loop
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				actions := ds.Tick()
				for _, act := range actions {
					switch act.Type {
					case "assign":
						if err := mgr.AssignToTable(act.BotID, act.TableID); err != nil {
							logg.Error("assign failed", "bot", act.BotID, "table", act.TableID, "error", err)
						} else {
							logg.Info("assigned bot", "bot", act.BotID, "table", act.TableID)
						}
					case "unassign":
						if err := mgr.UnassignFromTable(act.BotID); err != nil {
							logg.Error("unassign failed", "bot", act.BotID, "error", err)
						} else {
							logg.Info("unassigned bot", "bot", act.BotID)
						}
					}
				}
			case <-stopCh:
				return
			}
		}
	}()

	// Step 7: Watch hand results to increment hand count in scheduler
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for botID, ab := range ds.ActiveBots() {
					hands := mgr.GetBotHandsPlayed(botID)
					for i := ab.HandsPlayed; i < hands; i++ {
						ds.RecordHandPlayed(botID)
					}
				}
			case <-stopCh:
				return
			}
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logg.Info("shutting down")
	close(stopCh)
	mgr.StopAll()
	logg.Info("done")
}
