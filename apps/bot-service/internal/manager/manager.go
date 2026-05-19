package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/depuzhiguang/bot-service/internal/ai"
	"github.com/depuzhiguang/bot-service/internal/identity"
)

type Bot struct {
	Profile identity.Profile
	Engine  *ai.Engine
	TableID string
	Status  string
	ctx     context.Context
	cancel  context.CancelFunc
}

type Manager struct {
	bots   map[string]*Bot
	mu     sync.RWMutex
	tables map[string][]string
}

func NewManager() *Manager {
	return &Manager{
		bots:   make(map[string]*Bot),
		tables: make(map[string][]string),
	}
}

func (m *Manager) Spawn(count int) {
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

func (m *Manager) AssignToTable(botID, tableID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	bot, ok := m.bots[botID]
	if !ok {
		return fmt.Errorf("bot not found: %s", botID)
	}
	bot.TableID = tableID
	bot.Status = "playing"
	m.tables[tableID] = append(m.tables[tableID], botID)
	go m.runBot(bot)
	return nil
}

func (m *Manager) runBot(bot *Bot) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-bot.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, bot := range m.bots {
		bot.cancel()
	}
	m.bots = make(map[string]*Bot)
	m.tables = make(map[string][]string)
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
