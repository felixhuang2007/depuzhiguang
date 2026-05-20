package registrar

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateProfile(t *testing.T) {
	p := GenerateProfile(0, "tight_aggressive")
	assert.NotEmpty(t, p.Username)
	assert.NotEmpty(t, p.Password)
	assert.Equal(t, "tight_aggressive", p.Style)
	assert.Equal(t, 10000, p.InitialGold)
	assert.True(t, len(p.Password) >= 8)
}

func TestGenerateProfile_UniqueUsername(t *testing.T) {
	p1 := GenerateProfile(0, "nit")
	p2 := GenerateProfile(1, "nit")
	assert.NotEqual(t, p1.Username, p2.Username)
}

func TestRegistrar_RegisterUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/register", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"user-123","accessToken":"abc123"}`))
	}))
	defer srv.Close()

	reg := NewRegistrar(srv.URL)
	userID, token, err := reg.RegisterUser(SimProfile{Username: "sim_test", Email: "t@test.com", Password: "pass", Nickname: "Test"})
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)
	assert.Equal(t, "abc123", token)
}

func TestRegistrar_RegisterUser_ConflictFallback(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/auth/register" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if r.URL.Path == "/api/auth/login" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"accessToken":"fallback_token","user":{"id":"user-fb"}}`))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	}))
	defer srv.Close()

	reg := NewRegistrar(srv.URL)
	userID, token, err := reg.RegisterUser(SimProfile{Username: "sim_test", Email: "t@test.com", Password: "pass", Nickname: "Test"})
	require.NoError(t, err)
	assert.Equal(t, "user-fb", userID)
	assert.Equal(t, "fallback_token", token)
	assert.Equal(t, 2, callCount)
}

func TestRegistrar_RegisterUser_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	reg := NewRegistrar(srv.URL)
	_, _, err := reg.RegisterUser(SimProfile{Username: "sim_test", Email: "t@test.com", Password: "pass", Nickname: "Test"})
	require.Error(t, err)
}

func TestRegistrar_RegisterBatch(t *testing.T) {
	registered := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		registered++
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":"user-%d","accessToken":"tok"}`, registered)
	}))
	defer srv.Close()

	reg := NewRegistrar(srv.URL)
	profiles, tokens, err := reg.RegisterBatch(4)
	require.NoError(t, err)
	assert.Len(t, profiles, 4)
	assert.Len(t, tokens, 4)
	assert.Equal(t, 4, registered)
	for _, p := range profiles {
		assert.NotEmpty(t, p.UserID)
	}
}
