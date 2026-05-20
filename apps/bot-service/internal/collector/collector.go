package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Collector struct {
	apiBaseURL string
	client     *http.Client
}

func NewCollector(apiBaseURL string) *Collector {
	return &Collector{
		apiBaseURL: apiBaseURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

type ActionRecord struct {
	SessionID   string `json:"sessionId"`
	UserID      string `json:"userId"`
	TableID     string `json:"tableId"`
	HandNumber  int    `json:"handNumber"`
	Phase       string `json:"phase"`
	Action      string `json:"action"`
	Amount      int    `json:"amount"`
	PotBefore   int    `json:"potBefore"`
	PotAfter    int    `json:"potAfter"`
	StackBefore int    `json:"stackBefore"`
	StackAfter  int    `json:"stackAfter"`
	HoleCards   string `json:"holeCards,omitempty"`
	Community   string `json:"community,omitempty"`
}

func (c *Collector) LogAction(record ActionRecord) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal action: %w", err)
	}
	resp, err := c.client.Post(c.apiBaseURL+"/api/sim/actions", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("log action failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("log action returned %d", resp.StatusCode)
	}
	return nil
}

type HandResult struct {
	SessionID string  `json:"sessionId"`
	UserID    string  `json:"userId"`
	TableID   string  `json:"tableId"`
	WinAmount int     `json:"winAmount"`
	IsWinner  bool    `json:"isWinner"`
	BBWon     float64 `json:"bbWon"`
}

func (c *Collector) RecordResult(result HandResult) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	resp, err := c.client.Post(c.apiBaseURL+"/api/sim/results", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("record result failed: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
