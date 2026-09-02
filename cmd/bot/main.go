package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"mumblebeats/internal/api"
	"mumblebeats/internal/audio"
	"mumblebeats/internal/config"
	"mumblebeats/internal/db"
	"mumblebeats/internal/mumble"
)

func main() {
	if handleSelfMove() {
		return
	}

	// Configurar log a archivo para poder depurar cuando usamos -H=windowsgui
	logFile, err := os.OpenFile("mumblebeats.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		os.Stdout = logFile
		os.Stderr = logFile
	}
	fmt.Println("--- Iniciando MumbleBeats ---")
	fmt.Println("Fecha:", time.Now().Format(time.RFC3339))

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
		fmt.Printf("ADVERTENCIA: No se pudieron descargar las dependencias automáticamente: %v\n", err)
		fmt.Println("Las funciones de YouTube podrían no funcionar. Puedes descargar yt-dlp.exe y ffmpeg manualmente y ponerlos en esta carpeta.")
	} else {
		fmt.Println("Dependencias listas.")
	}

	// 3. Crear cliente de Mumble
	botClient := mumble.NewBotClient(cfg)

	// 4. Iniciar Servidor API Web PRIMERO
	fmt.Println("Iniciando API HTTP...")
	apiServer := api.NewServer(cfg, botClient)
	
	// Enganchar el reproductor con los WebSockets
	botClient.OnStateChange = func() {
		apiServer.Hub.Broadcast(apiServer.GetState())
	}
	
	go func() {
		if err := apiServer.Start(":8080"); err != nil {
			fmt.Printf("Error en Servidor HTTP: %v\n", err)
		}
	}()
	
	// Abrir navegador automáticamente
	openBrowser("http://localhost:8080")

	// 5. Iniciar Worker de la Cola dinámicamente
	fmt.Println("Iniciando el Queue Worker...")
	audio.StartQueueWorker(func() *audio.Player {
		return botClient.Player
	})

	// 6. Conectar a Mumble (No bloqueante / No exit fatal si falla)
	fmt.Println("Conectando al servidor de Mumble...")
	go func() {
		if err := botClient.Connect(); err != nil {
			fmt.Printf("ADVERTENCIA: Error al conectar a Mumble: %v\n", err)
			fmt.Println("Por favor, verifica la configuración en el panel web (http://localhost:8080).")
		}
	}()

	// 7. Mantener vivo el programa hasta que se reciba una señal de salida (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	fmt.Println("\nApagando MumbleBeats...")
	botClient.Disconnect()
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		fmt.Printf("No se pudo abrir el navegador automáticamente: %v\n", err)
	}
}

func handleSelfMove() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--cleanup-old=") {
			oldPath := strings.TrimPrefix(arg, "--cleanup-old=")
			go func() {
				time.Sleep(2 * time.Second)
				os.Remove(oldPath)
			}()
			return false
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	if strings.ToLower(filepath.Base(exeDir)) == "mumblebeats" {
		return false
	}

	if _, err := os.Stat(filepath.Join(exeDir, "config.json")); err == nil {
		return false
	}

	targetDir := filepath.Join(exeDir, "MumbleBeats")
	targetExe := filepath.Join(targetDir, exeName)

	if strings.EqualFold(exePath, targetExe) {
		return false
	}

	fmt.Println("Configurando entorno limpio en carpeta MumbleBeats...")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return false
	}

	src, err := os.Open(exePath)
	if err != nil {
		return false
	}
	dst, err := os.OpenFile(targetExe, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		src.Close()
		return false
	}
	_, err = io.Copy(dst, src)
	src.Close()
	dst.Close()

	if err != nil {
		return false
	}

	if runtime.GOOS != "windows" {
		os.Chmod(targetExe, 0755)
	}

	cmd := exec.Command(targetExe, fmt.Sprintf("--cleanup-old=%s", exePath))
	cmd.Dir = targetDir
	cmd.Start()

	os.Exit(0)
	return true
}
