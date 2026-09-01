package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"mumblebeats/internal/audio"
)

func main() {
	fmt.Println("Iniciando MumbleBeats...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Verificando dependencias (yt-dlp, ffmpeg)...")
	err := audio.EnsureDependencies(ctx)
	if err != nil {
		fmt.Printf("Error fatal al descargar dependencias: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Dependencias listas. Motor de audio preparado.")

	// Aquí irá la lógica de conexión a Mumble más adelante.
}
