package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mumblebeats/internal/api"
	"mumblebeats/internal/audio"
	"mumblebeats/internal/config"
	"mumblebeats/internal/db"
	"mumblebeats/internal/mumble"
)

func main() {
	fmt.Println("Iniciando MumbleBeats...")

	// 1. Cargar Configuración
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Error al cargar configuración: %v\n", err)
		os.Exit(1)
	}

	// 2. Inicializar Base de Datos SQLite
	fmt.Println("Inicializando base de datos SQLite...")
	if err := db.InitDB("mumblebeats.db"); err != nil {
		fmt.Printf("Error fatal al inicializar SQLite: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Indexando biblioteca local (carpeta 'music')...")
	if err := db.IndexLocalLibrary("music"); err != nil {
		fmt.Printf("Advertencia al indexar biblioteca local: %v\n", err)
	}

	// 3. Verificar Dependencias (yt-dlp, ffmpeg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Println("Verificando dependencias de audio...")
	if err := audio.EnsureDependencies(ctx); err != nil {
		fmt.Printf("Error fatal al descargar dependencias: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Dependencias listas.")

	// 3. Conectar a Mumble
	fmt.Println("Conectando al servidor de Mumble...")
	botClient := mumble.NewBotClient(cfg)
	if err := botClient.Connect(); err != nil {
		fmt.Printf("Error al conectar a Mumble: %v\n", err)
		os.Exit(1)
	}

	// 4. Iniciar Worker de la Cola
	fmt.Println("Iniciando el Queue Worker...")
	audio.StartQueueWorker(botClient.Player)

	// 5. Iniciar Servidor API Web
	fmt.Println("Iniciando API HTTP...")
	apiServer := api.NewServer(cfg, botClient)
	
	// Enganchar el reproductor con los WebSockets
	botClient.Player.OnStateChange = func() {
		apiServer.Hub.Broadcast(apiServer.GetState())
	}
	
	go func() {
		if err := apiServer.Start(":8080"); err != nil {
			fmt.Printf("Error en Servidor HTTP: %v\n", err)
		}
	}()

	// 5. Mantener vivo el programa hasta que se reciba una señal de salida (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	fmt.Println("\nApagando MumbleBeats...")
	if botClient.Client != nil {
		botClient.Client.Disconnect()
	}
}
