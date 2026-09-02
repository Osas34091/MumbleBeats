package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
}

func NewServer(cfg *config.Config, bot *mumble.BotClient) *Server {
	r := chi.NewRouter()
	
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
	}
	
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.Router.Route("/api", func(r chi.Router) {
		r.Get("/queue", s.handleGetQueue)
		r.Post("/play", s.handlePlay)
		r.Post("/skip", s.handleSkip)
		r.Post("/clear", s.handleClear)
		r.Post("/stop", s.handleStop)
		r.Post("/pause", s.handlePause)
		r.Post("/resume", s.handleResume)
		r.Delete("/queue/{id}", s.handleRemoveTrack)
		r.Post("/queue/{id}/up", s.handleMoveUp)
		r.Post("/queue/{id}/down", s.handleMoveDown)
		r.Get("/playlists", s.handleGetPlaylists)
		r.Post("/playlists/save", s.handleSavePlaylist)
		r.Post("/playlists/load", s.handleLoadPlaylist)
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

func (s *Server) handleGetQueue(w http.ResponseWriter, r *http.Request) {
	tracks, err := db.GetQueue(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Datos adicionales de Now Playing
	var currentTrack *db.Track
	position := 0
	isPaused := false
	speed := float32(1.0)
	
	if s.Bot.Player != nil && s.Bot.Player.CurrentTrack != nil {
		currentTrack = s.Bot.Player.CurrentTrack
		position = int(s.Bot.Player.Position.Seconds())
		isPaused = s.Bot.Player.IsPaused
		speed = s.Bot.Player.Speed
	}

	response := map[string]interface{}{
		"queue":       tracks,
		"now_playing": currentTrack,
		"position":    position,
		"is_paused":   isPaused,
		"speed":       speed,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	
	metadata, err := audio.FetchMetadata(req.Query)
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
