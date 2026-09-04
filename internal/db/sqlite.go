package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Track struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	AddedBy   string    `json:"added_by"`
	Thumbnail string    `json:"thumbnail"`
	Duration  float64   `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
}

var DB *sql.DB

func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite", filepath)
	if err != nil {
		return fmt.Errorf("error conectando a sqlite: %w", err)
	}
	
	// Prevenir error SQLITE_BUSY forzando a que sqlite trabaje en serie
	DB.SetMaxOpenConns(1)

	// Crear tabla de cola si no existe
	schema := `
	CREATE TABLE IF NOT EXISTS queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		added_by TEXT NOT NULL,
		thumbnail TEXT DEFAULT '',
		position INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS local_library (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL UNIQUE,
		path TEXT NOT NULL,
		indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS playlist_tracks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id INTEGER,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		type TEXT NOT NULL,
		thumbnail TEXT DEFAULT '',
		position INTEGER DEFAULT 0,
		FOREIGN KEY(playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id INTEGER,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	
	_, err = DB.Exec(schema)
	if err != nil {
		return err
	}
	
	// Si la columna position no existe, la añadimos (migración sencilla)
	DB.Exec("ALTER TABLE queue ADD COLUMN position INTEGER DEFAULT 0")

	// Migración para añadir duración si no existe
	DB.Exec("ALTER TABLE queue ADD COLUMN duration REAL DEFAULT 0;")

	// Poner todas las canciones en pending a failed por seguridad (si se cerró de golpe)
	_, _ = DB.Exec("UPDATE queue SET status = 'failed' WHERE status = 'playing'")

	return nil
}

func AddTrack(title, url, trackType, addedBy, thumbnail string, duration float64) (int64, error) {
	// Obtener la posición máxima actual
	var maxPos int
	DB.QueryRow("SELECT COALESCE(MAX(position), 0) FROM queue").Scan(&maxPos)
	newPos := maxPos + 1

	res, err := DB.Exec("INSERT INTO queue (title, url, type, status, added_by, thumbnail, position, duration) VALUES (?, ?, ?, 'pending', ?, ?, ?, ?)",
		title, url, trackType, addedBy, thumbnail, newPos, duration)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetNextPending() (*Track, error) {
	row := DB.QueryRow("SELECT id, title, url, type, status, added_by, thumbnail, duration, created_at FROM queue WHERE status = 'pending' ORDER BY position ASC, id ASC LIMIT 1")
	
	var track Track
	err := row.Scan(&track.ID, &track.Title, &track.URL, &track.Type, &track.Status, &track.AddedBy, &track.Thumbnail, &track.Duration, &track.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No hay más canciones
		}
		return nil, err
	}
	return &track, nil
}

func MarkPlaying(id int64) error {
	_, err := DB.Exec("UPDATE queue SET status = 'playing' WHERE id = ?", id)
	return err
}

func MarkPlayed(id int64) error {
	_, err := DB.Exec("UPDATE queue SET status = 'played' WHERE id = ?", id)
	return err
}

func GetQueue(limit int) ([]*Track, error) {
	rows, err := DB.Query("SELECT id, title, url, type, status, added_by, thumbnail, duration, created_at FROM queue WHERE status IN ('pending', 'playing') ORDER BY position ASC, id ASC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		var t Track
		if err := rows.Scan(&t.ID, &t.Title, &t.URL, &t.Type, &t.Status, &t.AddedBy, &t.Thumbnail, &t.Duration, &t.CreatedAt); err != nil {
			return nil, err
		}
		tracks = append(tracks, &t)
	}
	return tracks, nil
}

func RemoveTrack(id int) error {
	_, err := DB.Exec("DELETE FROM queue WHERE id = ?", id)
	return err
}

func MoveTrackUp(id int) error {
	// Obtener la posición actual
	var currentPos int
	err := DB.QueryRow("SELECT position FROM queue WHERE id = ?", id).Scan(&currentPos)
	if err != nil { return err }
	
	// Encontrar la canción inmediatamente anterior (posición menor) pero pendiente
	var prevID, prevPos int
	err = DB.QueryRow("SELECT id, position FROM queue WHERE status = 'pending' AND position < ? ORDER BY position DESC LIMIT 1", currentPos).Scan(&prevID, &prevPos)
	if err == sql.ErrNoRows { return nil } // Ya es la primera
	
	// Hacer swap
	tx, _ := DB.Begin()
	tx.Exec("UPDATE queue SET position = ? WHERE id = ?", prevPos, id)
	tx.Exec("UPDATE queue SET position = ? WHERE id = ?", currentPos, prevID)
	return tx.Commit()
}

func MoveTrackDown(id int) error {
	var currentPos int
	err := DB.QueryRow("SELECT position FROM queue WHERE id = ?", id).Scan(&currentPos)
	if err != nil { return err }
	
	var nextID, nextPos int
	err = DB.QueryRow("SELECT id, position FROM queue WHERE status = 'pending' AND position > ? ORDER BY position ASC LIMIT 1", currentPos).Scan(&nextID, &nextPos)
	if err == sql.ErrNoRows { return nil } // Ya es la última
	
	tx, _ := DB.Begin()
	tx.Exec("UPDATE queue SET position = ? WHERE id = ?", nextPos, id)
	tx.Exec("UPDATE queue SET position = ? WHERE id = ?", currentPos, nextID)
	return tx.Commit()
}

func SaveQueueToPlaylist(name string) error {
	tx, err := DB.Begin()
	if err != nil { return err }
	
	// Create playlist
	res, err := tx.Exec("INSERT INTO playlists (name) VALUES (?)", name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ya existe una playlist con ese nombre")
	}
	playlistID, _ := res.LastInsertId()
	
	// Copy tracks from queue
	rows, err := tx.Query("SELECT title, url, type, thumbnail FROM queue WHERE status IN ('pending', 'playing') ORDER BY position ASC, id ASC")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()
	
	pos := 1
	for rows.Next() {
		var title, url, trackType string
		var thumbnail sql.NullString
		if err := rows.Scan(&title, &url, &trackType, &thumbnail); err == nil {
			tx.Exec("INSERT INTO playlist_tracks (playlist_id, title, url, type, thumbnail, position) VALUES (?, ?, ?, ?, ?, ?)",
				playlistID, title, url, trackType, thumbnail.String, pos)
			pos++
		}
	}
	
	return tx.Commit()
}

func LoadPlaylist(name string, addedBy string) error {
	var playlistID int
	err := DB.QueryRow("SELECT id FROM playlists WHERE name = ?", name).Scan(&playlistID)
	if err != nil {
		return fmt.Errorf("no se encontró la playlist '%s'", name)
	}
	
	rows, err := DB.Query("SELECT title, url, type, thumbnail FROM playlist_tracks WHERE playlist_id = ? ORDER BY position ASC", playlistID)
	if err != nil { return err }
	
	type pTrack struct {
		title, url, trackType, thumbnail string
	}
	var tracks []pTrack
	
	for rows.Next() {
		var title, url, trackType string
		var thumbnail sql.NullString
		if err := rows.Scan(&title, &url, &trackType, &thumbnail); err == nil {
			tracks = append(tracks, pTrack{title, url, trackType, thumbnail.String})
		}
	}
	rows.Close() // Close BEFORE calling AddTrack to prevent SQLITE_BUSY
	
	for _, t := range tracks {
		_, err := AddTrack(t.title, t.url, t.trackType, addedBy, t.thumbnail)
		if err != nil {
			fmt.Printf("Error adding track from playlist: %v\n", err)
		}
	}
	return nil
}

func GetPlaylists() ([]string, error) {
	rows, err := DB.Query("SELECT name FROM playlists ORDER BY name ASC")
	if err != nil { return nil, err }
	defer rows.Close()
	var list []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		list = append(list, name)
	}
	return list, nil
}

func ClearQueue() error {
	_, err := DB.Exec("UPDATE queue SET status = 'played' WHERE status = 'pending'")
	return err
}

func IndexLocalLibrary(musicDir string) error {
	// Limpiar tabla actual para evitar duplicados si se borran archivos
	_, err := DB.Exec("DELETE FROM local_library")
	if err != nil {
		return err
	}

	stmt, err := DB.Prepare("INSERT INTO local_library (filename, path) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := os.Stat(musicDir); os.IsNotExist(err) {
		os.Mkdir(musicDir, 0755)
		return nil // Carpeta recién creada, está vacía
	}

	return filepath.Walk(musicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			_, err = stmt.Exec(info.Name(), path)
			if err != nil {
				log.Printf("Error indexando %s: %v", info.Name(), err)
			}
		}
		return nil
	})
}

func FindLocalFile(name string) (string, string, error) {
	query := `SELECT filename, path FROM local_library WHERE filename LIKE ? LIMIT 1`
	row := DB.QueryRow(query, "%"+name+"%")
	var filename, path string
	err := row.Scan(&filename, &path)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("no se encontró '%s' en la biblioteca", name)
		}
		return "", "", err
	}
	return path, filename, nil
}
