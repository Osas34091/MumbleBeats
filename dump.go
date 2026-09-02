package main
import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)
func main() {
	db, err := sql.Open("sqlite", "mumblebeats.db")
	if err != nil { panic(err) }
	rows, err := db.Query("SELECT id, thumbnail FROM queue")
	if err != nil { panic(err) }
	for rows.Next() {
		var id int
		var thumb string
		rows.Scan(&id, &thumb)
		fmt.Printf("ID %d: %s\n", id, thumb)
	}
}
