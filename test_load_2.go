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
	rows, _ := db.DB.Query("SELECT title, url, type, thumbnail FROM playlist_tracks WHERE playlist_id = ? ORDER BY position ASC", playlistID)
	defer rows.Close()
	
	for rows.Next() {
		var title, url, trackType string
		var thumbnail sql.NullString
		if err := rows.Scan(&title, &url, &trackType, &thumbnail); err == nil {
			id, err2 := db.AddTrack(title, url, trackType, "TestUser", thumbnail.String)
			fmt.Printf("AddTrack %s -> ID: %d, err: %v\n", title, id, err2)
		} else {
			fmt.Printf("Scan error: %v\n", err)
		}
	}
}
