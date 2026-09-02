package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	
	"mumblebeats/internal/audio"
	"mumblebeats/internal/config"
	"mumblebeats/internal/db"
	"mumblebeats/internal/mumble"
	"mumblebeats/web"
)

type Server struct {
	Config *config.Config
	Bot    *mumble.BotClient
	Router *chi.Mux
	Hub    *WSHub
}

func NewServer(cfg *config.Config, bot *mumble.BotClient) *Server {
	r := chi.NewRouter()
	
	hub := NewWSHub()
	go hub.Run()
	
	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // En producción, restringir al dominio de frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, 
	}))

	s := &Server{
		Config: cfg,
		Bot:    bot,
		Router: r,
		Hub:    hub,
	}
	
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.Router.Route("/api", func(r chi.Router) {
		r.Get("/ws", s.Hub.ServeWS)
		r.Get("/queue", s.handleGetQueue)
		r.Post("/play", s.handlePlay)
		r.Post("/skip", s.handleSkip)
		r.Post("/clear", s.handleClear)
		r.Post("/stop", s.handleStop)
		r.Post("/pause", s.handlePause)
		r.Post("/resume", s.handleResume)
		r.Post("/seek", s.handleSeek)
		r.Post("/speed", s.handleSpeed)
		r.Post("/filter", s.handleFilter)
		r.Post("/volume", s.handleVolume)
		r.Post("/shutdown", s.handleShutdown)
		r.Get("/config", s.handleGetConfig)
		r.Post("/config", s.handleSaveConfig)
		r.Delete("/queue/{id}", s.handleRemoveTrack)
		r.Post("/queue/{id}/up", s.handleMoveUp)
		r.Post("/queue/{id}/down", s.handleMoveDown)
		r.Get("/playlists", s.handleGetPlaylists)
		r.Post("/playlists/save", s.handleSavePlaylist)
		r.Post("/playlists/load", s.handleLoadPlaylist)
		r.Post("/upload", s.handleUpload)
	})
	
	// Servir frontend en la raíz usando go:embed
	distFS, err := web.DistFS()
	if err != nil {
		fmt.Printf("Advertencia: No se pudo cargar el frontend empaquetado: %v\n", err)
	} else {
		// Crear un FileServer para el sistema de archivos embebido
		fileServer := http.FileServer(http.FS(distFS))
		
		// Servir los estáticos bajo rutas específicas
		s.Router.Handle("/assets/*", fileServer)
		s.Router.Handle("/vite.svg", fileServer)
		
		// Redirigir cualquier otra ruta no "/api" al index.html
		s.Router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			file, err := distFS.Open("index.html")
			if err != nil {
				http.Error(w, "Index no encontrado", http.StatusInternalServerError)
				return
			}
			defer file.Close()
			stat, _ := file.Stat()
			http.ServeContent(w, r, "index.html", stat.ModTime(), file.(io.ReadSeeker))
		})
	}
}

func (s *Server) Start(port string) error {
	fmt.Printf("Servidor API Web iniciado en el puerto %s\n", port)
	return http.ListenAndServe(port, s.Router)
}

// Controladores

func (s *Server) GetState() map[string]interface{} {
	tracks, _ := db.GetQueue(50)
	
	var currentTrack *db.Track
	position := 0
	isPaused := false
	speed := float32(1.0)
	volume := float32(1.0)
	
	if s.Bot.Player != nil {
		speed = s.Bot.Player.Speed
		volume = s.Bot.Player.BaseVolume
		
		if s.Bot.Player.CurrentTrack != nil {
			currentTrack = s.Bot.Player.CurrentTrack
			position = int(s.Bot.Player.Position.Seconds())
			isPaused = s.Bot.Player.IsPaused
		}
	}

	return map[string]interface{}{
		"type":        "STATE_UPDATE",
		"queue":       tracks,
		"now_playing": currentTrack,
		"position":    position,
		"is_paused":   isPaused,
		"speed":       speed,
		"volume":      volume,
	}
}

func (s *Server) handleGetQueue(w http.ResponseWriter, r *http.Request) {
	state := s.GetState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

type PlayRequest struct {
	Query string `json:"query"`
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	var req PlayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	query := req.Query
	
	// 1. Es una URL?
	isURL := strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://")
	
	// 2. Si no es URL, buscar en archivos locales primero
	if !isURL {
		foundPath, foundName, err := db.FindLocalFile(query)
		if err == nil {
			// Encontrado localmente!
			id, addErr := db.AddTrack(foundName, foundPath, "local", "Dashboard", "")
			if addErr != nil {
				http.Error(w, addErr.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"id":      id,
				"track":   foundName,
			})
			return
		}
	}
	
	// 3. Fallback: yt-dlp (URL o Búsqueda en YouTube)
	metadata, err := audio.FetchMetadata(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := db.AddTrack(metadata.Title, metadata.WebpageURL, "youtube", "Dashboard", metadata.Thumbnail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"track":   metadata.Title,
	})
}

func (s *Server) handleSkip(w http.ResponseWriter, r *http.Request) {
	if s.Bot.Player != nil {
		s.Bot.Player.Stop()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleClear(w http.ResponseWriter, r *http.Request) {
	err := db.ClearQueue()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	db.ClearQueue()
	if s.Bot.Player != nil {
		s.Bot.Player.Stop()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	if s.Bot.Player != nil {
		s.Bot.Player.Pause()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleResume(w http.ResponseWriter, r *http.Request) {
	if s.Bot.Player != nil {
		s.Bot.Player.Resume()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleRemoveTrack(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	
	err := db.RemoveTrack(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleMoveUp(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	db.MoveTrackUp(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleMoveDown(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	db.MoveTrackDown(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleGetPlaylists(w http.ResponseWriter, r *http.Request) {
	lists, err := db.GetPlaylists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func (s *Server) handleSavePlaylist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := db.SaveQueueToPlaylist(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleLoadPlaylist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := db.LoadPlaylist(req.Name, "Dashboard")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// Limitar el tamaño de la subida a 50MB
	r.ParseMultipartForm(50 << 20)
	
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error recuperando archivo", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Crear directorio si no existe
	if err := os.MkdirAll("music", 0755); err != nil {
		http.Error(w, "Error creando directorio music", http.StatusInternalServerError)
		return
	}
	
	filePath := filepath.Join("music", handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error creando archivo", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error guardando archivo", http.StatusInternalServerError)
		return
	}
	
	// Reindexar base de datos
	db.IndexLocalLibrary("music")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Control PRO
func (s *Server) handleSeek(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Seconds int `json:"seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if s.Bot != nil && s.Bot.Player != nil {
		s.Bot.Player.Seek(req.Seconds)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSpeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Speed float64 `json:"speed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if s.Bot != nil && s.Bot.Player != nil {
		s.Bot.Player.SetSpeed(req.Speed)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleFilter(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filter string `json:"filter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if s.Bot != nil && s.Bot.Player != nil {
		s.Bot.Player.ApplyFilter(req.Filter)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleVolume(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Volume float32 `json:"volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if s.Bot != nil && s.Bot.Player != nil {
		s.Bot.Player.BaseVolume = req.Volume
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	go func() {
		time.Sleep(1 * time.Second)
		if s.Bot != nil && s.Bot.Client != nil {
			s.Bot.Client.Disconnect()
		}
		os.Exit(0)
	}()
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Bot.Config)
}

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Keep same password if masked or empty, just basic save
	if newCfg.Password == "" {
		newCfg.Password = s.Bot.Config.Password
	}
	
	*s.Bot.Config = newCfg
	err := config.SaveConfig(s.Bot.Config, "config.json")
	if err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}
	
	// Si la configuración se guarda correctamente, reconectamos en background
	go func() {
		s.Bot.Reconnect()
	}()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
