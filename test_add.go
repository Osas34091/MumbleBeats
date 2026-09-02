package main

import (
	"fmt"
	"mumblebeats/internal/db"
)

func main() {
	err := db.InitDB("mumblebeats.db")
	if err != nil {
		fmt.Println("Init DB Error:", err)
		return
	}
	
	err = db.LoadPlaylist("Prueba 1", "TestUser")
	if err != nil {
		fmt.Println("LoadPlaylist Error:", err)
	} else {
		fmt.Println("LoadPlaylist Success")
	}
	
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM queue").Scan(&count)
	fmt.Printf("Queue size: %d\n", count)
}
