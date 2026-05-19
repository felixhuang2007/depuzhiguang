package registrar

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type SimProfile struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Nickname    string `json:"nickname"`
	Style       string `json:"style"`
	InitialGold int    `json:"initialGold"`
}

type Registrar struct {
	apiBaseURL string
	client     *http.Client
}

func NewRegistrar(apiBaseURL string) *Registrar {
	return &Registrar{
		apiBaseURL: apiBaseURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func generateSecurePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	length := 12
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return "Sim" + string(b)
}

func GenerateProfile(index int, style string) SimProfile {
	firstNames := []string{"Tom", "Lily", "Jack", "Rose", "Alex", "Mia", "Leo", "Zoe", "Max", "Eva", "Sam", "Amy", "Ben", "Sara", "Dan", "Kim", "Ray", "Joy", "Jay", "Ann"}
	fn := firstNames[index%len(firstNames)]
	return SimProfile{
		Username:    fmt.Sprintf("sim_%s_%d", fn, index),
		Email:       fmt.Sprintf("sim_%s_%d@test.com", fn, index),
		Password:    generateSecurePassword(),
		Nickname:    fmt.Sprintf("%s the %s", fn, style),
		Style:       style,
		InitialGold: 10000,
	}
}

func (r *Registrar) RegisterUser(profile SimProfile) (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"username": profile.Username,
		"email":    profile.Email,
		"password": profile.Password,
		"nickname": profile.Nickname,
	})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := r.client.Post(r.apiBaseURL+"/api/auth/register", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("register request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return r.loginUser(profile)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("register returned status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode register response: %w", err)
	}
	return result.Token, nil
}

func (r *Registrar) loginUser(profile SimProfile) (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"username": profile.Username,
		"password": profile.Password,
	})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := r.client.Post(r.apiBaseURL+"/api/auth/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode login response: %w", err)
	}
	return result.AccessToken, nil
}

func (r *Registrar) RegisterBatch(count int) ([]SimProfile, []string, error) {
	personas := []string{
		"tight_aggressive", "loose_aggressive", "nit", "loose_passive",
		"maniac", "rock", "calling_station", "adaptive",
	}

	profiles := make([]SimProfile, count)
	tokens := make([]string, count)

	for i := 0; i < count; i++ {
		style := personas[i%len(personas)]
		profiles[i] = GenerateProfile(i, style)
		token, err := r.RegisterUser(profiles[i])
		if err != nil {
			return nil, nil, fmt.Errorf("register user %d: %w", i, err)
		}
		tokens[i] = token
		time.Sleep(100 * time.Millisecond) // Rate limit friendly
	}
	return profiles, tokens, nil
}
