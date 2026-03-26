package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/coder/websocket"
	"github.com/evcc-io/evcc/server/db"
	dbuser "github.com/evcc-io/evcc/server/db/user"
	"github.com/evcc-io/evcc/util/auth"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memSettings is an in-memory settings store for use in auth integration tests.
type memSettings struct {
	mu   sync.Mutex
	data map[string]string
}

func newMemSettings() *memSettings {
	return &memSettings{data: make(map[string]string)}
}

func (m *memSettings) String(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return v, nil
}

func (m *memSettings) SetString(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// setupTestDB initializes an in-memory SQLite database for tests.
// It creates a test admin user and returns the username and password.
func setupTestDB(t *testing.T) (username, password string) {
	t.Helper()
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	t.Cleanup(func() { db.Close() })

	username = "testadmin"
	password = "secret"

	u := dbuser.User{Username: username, Role: dbuser.RoleAdmin}
	require.NoError(t, u.SetPassword(password))
	require.NoError(t, db.Instance.Create(&u).Error)

	return username, password
}

// newAuthRouter builds a minimal router with auth middleware, /api/auth/* routes, a
// protected /api/protected endpoint, and the auth-protected /ws endpoint.
func newAuthRouter(t *testing.T, authObject auth.Auth) (*mux.Router, *SocketHub) {
	t.Helper()
	hub := NewSocketHub()

	router := mux.NewRouter()

	// unprotected auth routes
	apiAuth := router.PathPrefix("/api/auth").Subrouter()
	apiAuth.HandleFunc("/status", authStatusHandler(authObject)).Methods(http.MethodGet)
	apiAuth.HandleFunc("/login", loginHandler(authObject)).Methods(http.MethodPost)
	apiAuth.HandleFunc("/logout", logoutHandler).Methods(http.MethodPost)

	// auth-protected websocket
	router.Handle("/ws", ensureAuthHandler(authObject)(socketHandler(hub)))

	// auth-protected api
	api := router.PathPrefix("/api").Subrouter()
	api.Use(ensureAuthHandler(authObject))
	api.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	return router, hub
}

// loginAndGetCookie performs a login request and returns the auth cookie.
func loginAndGetCookie(t *testing.T, srv *httptest.Server, username, password string) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(loginRequest{Username: username, Password: password})
	resp, err := http.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	for _, c := range resp.Cookies() {
		if c.Name == authCookieName {
			return c
		}
	}
	t.Fatal("auth cookie not found in login response")
	return nil
}

func TestEnsureAuthRejectsUnauthenticated(t *testing.T) {
	setupTestDB(t)
	authObject := auth.NewMock(newMemSettings())
	router, _ := newAuthRouter(t, authObject)
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/protected")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestEnsureAuthAllowsAfterLogin(t *testing.T) {
	username, password := setupTestDB(t)
	authObject := auth.NewMock(newMemSettings())
	router, _ := newAuthRouter(t, authObject)
	srv := httptest.NewServer(router)
	defer srv.Close()

	cookie := loginAndGetCookie(t, srv, username, password)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/protected", nil)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthStatusRouteAccessibleWithoutLogin(t *testing.T) {
	setupTestDB(t)
	authObject := auth.NewMock(newMemSettings())
	router, _ := newAuthRouter(t, authObject)
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/auth/status")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	username, _ := setupTestDB(t)
	authObject := auth.NewMock(newMemSettings())
	router, _ := newAuthRouter(t, authObject)
	srv := httptest.NewServer(router)
	defer srv.Close()

	body, _ := json.Marshal(loginRequest{Username: username, Password: "wrong"})
	resp, err := http.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocketRejectsUnauthenticated(t *testing.T) {
	setupTestDB(t)
	authObject := auth.NewMock(newMemSettings())
	router, _ := newAuthRouter(t, authObject)
	srv := httptest.NewServer(router)
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/ws"
	_, _, err := websocket.Dial(t.Context(), wsURL, nil)
	require.Error(t, err, "unauthenticated WebSocket connection should be rejected")
}

func TestWebSocketConnectsAfterLogin(t *testing.T) {
	username, password := setupTestDB(t)
	authObject := auth.NewMock(newMemSettings())
	router, hub := newAuthRouter(t, authObject)
	srv := httptest.NewServer(router)
	defer srv.Close()

	// Drain the hub's register channel so subscribe() doesn't block after
	// the welcome send; the hub's Run loop is not started in tests.
	go func() {
		for s := range hub.register {
			s.send <- []byte("{}")
		}
	}()

	cookie := loginAndGetCookie(t, srv, username, password)

	wsURL := "ws" + srv.URL[len("http"):] + "/ws"
	conn, _, err := websocket.Dial(t.Context(), wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Cookie": {cookie.String()}},
	})
	require.NoError(t, err, "authenticated WebSocket connection should succeed")
	conn.CloseNow()
}
