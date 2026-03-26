package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	dbuser "github.com/evcc-io/evcc/server/db/user"
	"github.com/evcc-io/evcc/util/auth"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const authCookieName = "auth"

type contextKey string

const (
	contextKeyUsername contextKey = "username"
	contextKeyRole     contextKey = "role"
)

type updatePasswordRequest struct {
	Current string `json:"current"`
	New     string `json:"new"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// isFirstSetup returns true when no users exist in the database yet
func isFirstSetup() bool {
	count, err := dbuser.Count()
	return err != nil || count == 0
}

func updatePasswordHandler(authObject auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authObject.GetAuthMode() == auth.Locked {
			http.Error(w, "Forbidden in demo mode", http.StatusForbidden)
			return
		}

		var req updatePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get calling user from context
		username, _ := r.Context().Value(contextKeyUsername).(string)
		if username == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		u, err := dbuser.ByUsername(username)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if !u.CheckPassword(req.Current) {
			http.Error(w, "Invalid password", http.StatusBadRequest)
			return
		}

		if err := u.SetPassword(req.New); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := dbuser.Save(u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Renew the auth cookie so the session stays valid after the password change.
		if err := setAuthCookieForUser(authObject, u.Username, string(u.Role), w); err != nil {
			http.Error(w, "Failed to renew JWT token.", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

// read jwt from header and cookie
func jwtFromRequest(r *http.Request) string {
	// read from header
	authHeader := r.Header.Get("Authorization")
	if token, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		return token
	}

	// read from cookie
	if cookie, _ := r.Cookie(authCookieName); cookie != nil {
		return cookie.Value
	}

	return ""
}

// authStatusHandler login status (true/false) based on jwt token. Error if no users configured.
func authStatusHandler(authObject auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authObject.GetAuthMode() == auth.Disabled {
			w.Write([]byte("true"))
			return
		}

		if authObject.GetAuthMode() == auth.Locked {
			http.Error(w, "Forbidden in demo mode", http.StatusForbidden)
			return
		}

		if isFirstSetup() {
			http.Error(w, "Not implemented", http.StatusNotImplemented)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		claims, err := authObject.ValidateJwtToken(jwtFromRequest(r))
		if err != nil || claims == nil {
			w.Write([]byte("false"))
			return
		}
		w.Write([]byte("true"))
	}
}

func setAuthCookieForUser(authObject auth.Auth, username, role string, w http.ResponseWriter) error {
	lifetime := time.Hour * 24 * 90 // 90 day valid
	tokenString, err := authObject.GenerateJwtToken(username, role, lifetime)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(lifetime),
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}

func loginHandler(authObject auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authObject.GetAuthMode() == auth.Locked {
			http.Error(w, "Forbidden in demo mode", http.StatusForbidden)
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err := dbuser.ByUsername(req.Username)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !u.CheckPassword(req.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := setAuthCookieForUser(authObject, u.Username, string(u.Role), w); err != nil {
			http.Error(w, "Failed to generate JWT token.", http.StatusInternalServerError)
			return
		}
	}
}

// setupHandler handles the initial admin user creation. Only works when no users exist.
func setupHandler(authObject auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authObject.GetAuthMode() == auth.Locked {
			http.Error(w, "Forbidden in demo mode", http.StatusForbidden)
			return
		}

		if !isFirstSetup() {
			http.Error(w, "Setup already completed", http.StatusConflict)
			return
		}

		var req setupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" {
			http.Error(w, "Username and password are required", http.StatusBadRequest)
			return
		}

		u := dbuser.User{
			Username: req.Username,
			Role:     dbuser.RoleAdmin,
		}
		if err := u.SetPassword(req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := dbuser.Create(&u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := setAuthCookieForUser(authObject, u.Username, string(u.Role), w); err != nil {
			http.Error(w, "Failed to generate JWT token.", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Path:     "/",
		HttpOnly: true,
	})
}

// roleLevels maps each role to a numeric level for comparison.
var roleLevels = map[dbuser.Role]int{
	dbuser.RoleViewer:     0,
	dbuser.RoleUser:       1,
	dbuser.RoleMaintainer: 2,
	dbuser.RoleAdmin:      3,
}

// isRequestAllowed returns true when the given role is permitted to make the
// request. Rules are evaluated top-to-bottom; the first match wins.
func isRequestAllowed(role dbuser.Role, path, method string) bool {
	level := roleLevels[role]

	switch {
	// User management — admin only
	case strings.HasPrefix(path, "/api/users"):
		return level >= roleLevels[dbuser.RoleAdmin]

	// Destructive system ops — admin only
	case path == "/api/system/backup" || path == "/api/system/restore" ||
		path == "/api/system/reset" || path == "/api/system/shutdown" ||
		path == "/api/system/cache":
		return level >= roleLevels[dbuser.RoleAdmin]

	// Config & system (log, etc.) — maintainer+
	case strings.HasPrefix(path, "/api/config/") ||
		strings.HasPrefix(path, "/api/system/"):
		return level >= roleLevels[dbuser.RoleMaintainer]

	// Safe reads — viewer+
	case method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions:
		return level >= roleLevels[dbuser.RoleViewer]

	// All other writes (POST/PUT/DELETE/PATCH) on site routes — user+
	default:
		return level >= roleLevels[dbuser.RoleUser]
	}
}

func ensureAuthHandler(authObject auth.Auth) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authObject.GetAuthMode() == auth.Disabled {
				ctx := context.WithValue(r.Context(), contextKeyRole, string(dbuser.RoleAdmin))
				ctx = context.WithValue(ctx, contextKeyUsername, "admin")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if authObject.GetAuthMode() == auth.Locked {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := authObject.ValidateJwtToken(jwtFromRequest(r))
			if err != nil || claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !isRequestAllowed(dbuser.Role(claims.Role), r.URL.Path, r.Method) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUsername, claims.Subject)
			ctx = context.WithValue(ctx, contextKeyRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getUserByID is a helper for user handlers to look up a user by the {id} path variable
func getUserByID(r *http.Request) (*dbuser.User, error) {
	return dbuser.ByID(mux.Vars(r)["id"])
}

// isLastAdmin returns true if the given user is the last admin in the system
func isLastAdmin(userID uint) bool {
	count, err := dbuser.AdminCount(userID)
	return err != nil || count == 0
}

// notFound is a helper to check for gorm not found errors
func notFound(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
}
