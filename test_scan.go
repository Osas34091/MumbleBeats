package main

import (
	"fmt"
	"database/sql"
	"mumblebeats/internal/db"
)

func main() {
	err := db.InitDB("mumblebeats.db")
	if err != nil {
		fmt.Println("Init DB Error:", err)
		return
	}
	
	playlistID := 3 // Prueba 1
	
	rows, err := db.DB.Query("SELECT title, url, type, thumbnail FROM playlist_tracks WHERE playlist_id = ? ORDER BY position ASC", playlistID)
	if err != nil { 
		fmt.Println("Query error:", err)
		return 
	}
	defer rows.Close()
	
	for rows.Next() {
		var title, url, trackType string
		var thumbnail sql.NullString
		if err := rows.Scan(&title, &url, &trackType, &thumbnail); err != nil {
			fmt.Printf("Scan error: %v\n", err)
		} else {
			fmt.Printf("Scanned OK: %s\n", title)
		}
	}
}
