package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/depuzhiguang/bot-service/internal/ai"
	"github.com/depuzhiguang/bot-service/internal/client"
	"github.com/depuzhiguang/bot-service/internal/collector"
	"github.com/depuzhiguang/bot-service/internal/identity"
)

type Bot struct {
	Profile identity.Profile
	Engine  *ai.Engine
	TableID string
	Status  string
	client  *client.GameClient
	ctx     context.Context
	cancel  context.CancelFunc
}

type Manager struct {
	bots      map[string]*Bot
	mu        sync.RWMutex
	tables    map[string][]string
	wsURL     string
	apiURL    string
	tokens    map[string]string
	collector *collector.Collector
}

func NewManager(wsURL, apiURL string, tokens map[string]string) *Manager {
	return &Manager{
		bots:      make(map[string]*Bot),
		tables:    make(map[string][]string),
		wsURL:     wsURL,
		apiURL:    apiURL,
		tokens:    tokens,
		collector: collector.NewCollector(apiURL),
	}
}

func (m *Manager) Spawn(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < count; i++ {
		profile := identity.GenerateProfile(i)
		diff := ai.Regular
		switch i % 20 {
		case 0, 1, 2, 3, 4, 5:
			diff = ai.Fish
		case 16, 17, 18:
			diff = ai.Shark
		case 19:
			diff = ai.Whale
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.bots[profile.ID] = &Bot{
			Profile: profile,
			Engine:  ai.NewEngine(diff, "BTN"),
			Status:  "idle",
			ctx:     ctx,
			cancel:  cancel,
		}
	}
}

func (m *Manager) RegisterBot(botID string, engine *ai.Engine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	m.bots[botID] = &Bot{
		Profile: identity.Profile{ID: botID},
		Engine:  engine,
		Status:  "idle",
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (m *Manager) AssignToTable(botID, tableID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	bot, ok := m.bots[botID]
	if !ok {
		return fmt.Errorf("bot not found: %s", botID)
	}

	// Call REST API to join table first (creates TablePlayer record)
	token, hasToken := m.tokens[botID]
	if hasToken {
		payload, _ := json.Marshal(map[string]interface{}{"tableId": tableID})
		req, _ := http.NewRequest("POST", m.apiURL+"/api/tables/"+tableID+"/join", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[%s] Join table API call failed: %v", botID, err)
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusInternalServerError {
				// Likely insufficient gold — trigger refill and retry once
				log.Printf("[%s] Join returned 500, attempting gold refill", botID)
				refillPayload, _ := json.Marshal(map[string]interface{}{"user_id": botID})
				refillReq, _ := http.NewRequest("POST", m.apiURL+"/api/sim/refill", bytes.NewReader(refillPayload))
				refillReq.Header.Set("Content-Type", "application/json")
				if refillResp, refillErr := client.Do(refillReq); refillErr == nil {
					refillResp.Body.Close()
					log.Printf("[%s] Gold refill attempted, status %d", botID, refillResp.StatusCode)
					// Retry join
					retryReq, _ := http.NewRequest("POST", m.apiURL+"/api/tables/"+tableID+"/join", bytes.NewReader(payload))
					retryReq.Header.Set("Content-Type", "application/json")
					retryReq.Header.Set("Authorization", "Bearer "+token)
					if retryResp, retryErr := client.Do(retryReq); retryErr == nil {
						retryResp.Body.Close()
						log.Printf("[%s] Retry join status %d", botID, retryResp.StatusCode)
					}
				}
			} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				log.Printf("[%s] Join table API returned status %d", botID, resp.StatusCode)
			}
		}
	}

	bot.TableID = tableID
	bot.Status = "playing"

	// Check for duplicates before appending
	for _, existingID := range m.tables[tableID] {
		if existingID == botID {
			return nil // already assigned
		}
	}
	m.tables[tableID] = append(m.tables[tableID], botID)

	// Create and connect WebSocket client
	gc := client.NewGameClient(m.wsURL, bot.Profile.ID, tableID, bot.Engine)
	gc.SetActionCallback(func(phase, action string, amount, pot, stack int) {
		if err := m.collector.LogAction(collector.ActionRecord{
			UserID:      bot.Profile.ID,
			TableID:     tableID,
			Phase:       phase,
			Action:      action,
			Amount:      amount,
			PotBefore:   pot,
			StackBefore: stack,
		}); err != nil {
			log.Printf("[%s] Failed to log action: %v", bot.Profile.ID, err)
		}
	})
	bot.client = gc
	if err := gc.Connect(); err != nil {
		bot.Status = "error"
		return fmt.Errorf("bot connect failed: %w", err)
	}

	go m.runBot(bot)
	return nil
}

func (m *Manager) runBot(bot *Bot) {
	if bot.client != nil {
		bot.client.Run()
	}
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, bot := range m.bots {
		if bot.client != nil {
			bot.client.Stop()
		}
		bot.cancel()
	}
	m.bots = make(map[string]*Bot)
	m.tables = make(map[string][]string)
}

func (m *Manager) BotIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.bots))
	for id := range m.bots {
		ids = append(ids, id)
	}
	return ids
}

func (m *Manager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total, playing := len(m.bots), 0
	for _, bot := range m.bots {
		if bot.Status == "playing" {
			playing++
		}
	}
	return map[string]interface{}{"total": total, "playing": playing, "tables": len(m.tables)}
}

func (m *Manager) UnassignFromTable(botID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bot, ok := m.bots[botID]
	if !ok {
		return fmt.Errorf("bot not found: %s", botID)
	}

	tableID := bot.TableID

	// Call REST API to leave table first (returns gold)
	if tableID != "" {
		token, hasToken := m.tokens[botID]
		if hasToken {
			req, _ := http.NewRequest("POST", m.apiURL+"/api/tables/"+tableID+"/leave", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("[%s] Leave table API call failed: %v", botID, err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					log.Printf("[%s] Leave table API returned status %d", botID, resp.StatusCode)
				}
			}
		}
	}

	if bot.client != nil {
		bot.client.Leave()
	}

	bot.Status = "idle"
	bot.TableID = ""

	// Remove from table assignment
	for tid, ids := range m.tables {
		filtered := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != botID {
				filtered = append(filtered, id)
			}
		}
		m.tables[tid] = filtered
	}

	return nil
}

func (m *Manager) GetBot(botID string) (*Bot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bot, ok := m.bots[botID]
	return bot, ok
}

func (m *Manager) GetTableBots(tableID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, len(m.tables[tableID]))
	copy(ids, m.tables[tableID])
	return ids
}

func (m *Manager) GetBotHandsPlayed(botID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bot, ok := m.bots[botID]
	if !ok || bot.client == nil {
		return 0
	}
	return bot.client.HandsPlayed()
}
