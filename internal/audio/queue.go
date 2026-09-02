package audio

import (
	"fmt"
	"time"

	"mumblebeats/internal/db"
)

// StartQueueWorker inicia un bucle infinito en segundo plano que vigila la cola
func StartQueueWorker(player *Player) {
	go func() {
		for {
			// Buscar la siguiente canción en la cola
			track, err := db.GetNextPending()
			if err != nil {
				fmt.Printf("Error obteniendo siguiente track: %v\n", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Si no hay canciones, esperamos un poco y volvemos a comprobar
			if track == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			// Marcar como reproduciendo
			db.MarkPlaying(track.ID)

			var playErr error
			switch track.Type {
			case "youtube":
				playErr = player.PlayURL(track)
			case "radio":
				playErr = player.PlayRadio(track)
			case "local":
				playErr = player.PlayLocal(track)
			}

			if playErr != nil {
				fmt.Printf("Error reproduciendo track ID %d: %v\n", track.ID, playErr)
				if player.client.Self.Channel != nil {
					player.client.Self.Channel.Send(fmt.Sprintf("Error reproduciendo track: %v", playErr), false)
				}
			}

			// Marcar como reproducida al terminar (ya sea porque terminó sola o fue saltada)
			db.MarkPlayed(track.ID)
		}
	}()
}
