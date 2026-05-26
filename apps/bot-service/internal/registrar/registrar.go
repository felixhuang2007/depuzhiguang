package registrar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SimProfile struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Nickname    string `json:"nickname"`
	Style       string `json:"style"`
	InitialGold int    `json:"initialGold"`
	Token       string `json:"-"`
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

func generateSecurePassword(username string) string {
	// Deterministic password per username so restart can login after 409
	return "SimPwd_" + username + "_2025"
}

func GenerateProfile(index int, style string) SimProfile {
	firstNames := []string{"Tom", "Lily", "Jack", "Rose", "Alex", "Mia", "Leo", "Zoe", "Max", "Eva", "Sam", "Amy", "Ben", "Sara", "Dan", "Kim", "Ray", "Joy", "Jay", "Ann"}
	fn := firstNames[index%len(firstNames)]
	return SimProfile{
		Username:    fmt.Sprintf("sim_%s_%d", fn, index),
		Email:       fmt.Sprintf("sim_%s_%d@test.com", fn, index),
		Password:    generateSecurePassword(fmt.Sprintf("sim_%s_%d", fn, index)),
		Nickname:    fmt.Sprintf("%s the %s", fn, style),
		Style:       style,
		InitialGold: 10000,
	}
}

func (r *Registrar) RegisterUser(profile SimProfile) (string, string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"username":    profile.Username,
		"email":       profile.Email,
		"password":    profile.Password,
		"nickname":    profile.Nickname,
		"isSimUser":   true,
		"simStyle":    profile.Style,
		"initialGold": profile.InitialGold,
	})
	if err != nil {
		return "", "", fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := r.client.Post(r.apiBaseURL+"/api/auth/register", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", "", fmt.Errorf("register request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return r.loginUser(profile)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("register returned status %d", resp.StatusCode)
	}

	var result struct {
		ID          string `json:"id"`
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode register response: %w", err)
	}
	return result.ID, result.AccessToken, nil
}

func (r *Registrar) loginUser(profile SimProfile) (string, string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"username": profile.Username,
		"password": profile.Password,
	})
	if err != nil {
		return "", "", fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := r.client.Post(r.apiBaseURL+"/api/auth/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("login returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode login response: %w", err)
	}
	return result.User.ID, result.AccessToken, nil
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
		userID, token, err := r.RegisterUser(profiles[i])
		if err != nil {
			return nil, nil, fmt.Errorf("register user %d: %w", i, err)
		}
		profiles[i].UserID = userID
		profiles[i].Token = token
		tokens[i] = token
		time.Sleep(100 * time.Millisecond) // Rate limit friendly
	}
	return profiles, tokens, nil
}
