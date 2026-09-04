package api

import (
	"encoding/json"
	"net/http"
	"time"

	"mumblebeats/internal/config"
	"mumblebeats/internal/db"
)

func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	hasAdmin, err := db.HasAdmin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"is_setup": hasAdmin})
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	hasAdmin, err := db.HasAdmin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if hasAdmin {
		http.Error(w, "Setup ya fue completado", http.StatusForbidden)
		return
	}

	var req struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		MumbleAddress  string `json:"mumble_address"`
		MumblePort     string `json:"mumble_port"`
		MumbleUsername string `json:"mumble_username"`
		MumblePassword string `json:"mumble_password"`
		MumbleChannel  string `json:"mumble_channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Usuario y contraseña son requeridos", http.StatusBadRequest)
		return
	}

	// Create user
	userID, err := db.CreateUser(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update Mumble config
	s.Bot.Config.ServerAddress = req.MumbleAddress
	s.Bot.Config.ServerPort = req.MumblePort
	s.Bot.Config.Username = req.MumbleUsername
	s.Bot.Config.Password = req.MumblePassword
	s.Bot.Config.ChannelID = 0

	config.SaveConfig(s.Bot.Config, "config.json")
	
	// Start bot connection in background if it isn't running
	go func() {
		if err := s.Bot.Reconnect(); err != nil {
			// Will just log internally
		}
	}()

	// Create session and set cookie
	token, err := db.CreateSession(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	http.SetCookie(w, &http.Cookie{
		Name:     "mumblebeats_session",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := db.VerifyLogin(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	token, err := db.CreateSession(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mumblebeats_session",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("mumblebeats_session")
	if err == nil {
		db.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mumblebeats_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("mumblebeats_session")
	if err != nil {
		http.Error(w, "No autenticado", http.StatusUnauthorized)
		return
	}

	username, err := db.GetUserFromSession(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}
