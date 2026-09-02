package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync/atomic"
	"time"

	"mumblebeats/internal/db"

	"layeh.com/gumble/gumble"
)

type Player struct {
	client       *gumble.Client
	cancel       context.CancelFunc
	duckingCount atomic.Int32
	BaseVolume   float32
	DuckingVol   float32
	
	// Control de Reproducción
	CurrentTrack *db.Track
	StreamURL    string
	StartTime    time.Duration
	Position     time.Duration
	Speed        float32
	IsPaused     bool
	
	OnStateChange func() // Callback para notificar cambios (WebSockets)
}

func NewPlayer(client *gumble.Client) *Player {
	return &Player{
		client:     client,
		BaseVolume: 1.0,
		DuckingVol: 0.2, // 20% volume when ducking
		Speed:      1.0,
	}
}

// PlayURL extrae la URL con yt-dlp y reproduce el audio
func (p *Player) PlayURL(track *db.Track) error {
	p.Stop() // Detener lo que esté sonando
	
	p.CurrentTrack = track
	p.Position = 0
	p.IsPaused = false

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	fmt.Println("Extrayendo URL de audio con yt-dlp...")
	
	exeYtDlp := ResolveExecutable("yt-dlp")

	// Obtener la mejor URL de audio y metadata en formato JSON, ignorando playlists
	cmdYt := exec.CommandContext(ctx, exeYtDlp, "--no-playlist", "-J", "-f", "bestaudio", track.URL)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmdYt.Stdout = &out
	cmdYt.Stderr = &errOut
	if err := cmdYt.Run(); err != nil {
		return fmt.Errorf("error yt-dlp: %v, stderr: %s", err, errOut.String())
	}

	var metadata struct {
		Title      string `json:"title"`
		Uploader   string `json:"uploader"`
		Thumbnail  string `json:"thumbnail"`
		WebpageURL string `json:"webpage_url"`
		URL        string `json:"url"`
	}

	if err := json.Unmarshal(out.Bytes(), &metadata); err != nil {
		return fmt.Errorf("error parseando JSON de yt-dlp: %v", err)
	}

	streamURL := metadata.URL
	if streamURL == "" {
		return fmt.Errorf("yt-dlp no devolvió ninguna URL de stream")
	}

	// Enviar mensaje HTML al canal si estamos conectados a uno
	if p.client.Self.Channel != nil {
		var imgTag string
		imgBase64 := GetThumbnailBase64(metadata.Thumbnail, "mqdefault")
		if imgBase64 != "" {
			imgTag = fmt.Sprintf(`<br/><br/><img src="%s" height="90" />`, imgBase64)
		}

		msg := fmt.Sprintf(`Reproduciendo ahora: <a href="%s"><b>%s</b></a> por <b>%s</b> (Pedido por %s)%s`,
			metadata.WebpageURL, metadata.Title, metadata.Uploader, track.AddedBy, imgTag)
		p.client.Self.Channel.Send(msg, false)
	}

	fmt.Println("Iniciando FFmpeg...")
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
	
	return p.PlayDirectStream(ctx, streamURL)
}

// PlayRadio lanza ffmpeg directamente para reproducir una radio por internet
func (p *Player) PlayRadio(track *db.Track) error {
	p.Stop()
	
	p.CurrentTrack = track
	p.Position = 0
	p.IsPaused = false

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	if p.client.Self.Channel != nil {
		msg := fmt.Sprintf(`Reproduciendo Radio: <b>%s</b> (Añadido por %s)`, track.URL, track.AddedBy)
		p.client.Self.Channel.Send(msg, false)
	}
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}

	return p.PlayDirectStream(ctx, track.URL)
}



// PlayLocal reproduce un archivo local (la ruta debe ser exacta)
func (p *Player) PlayLocal(track *db.Track) error {
	p.Stop()
	
	p.CurrentTrack = track
	p.Position = 0
	p.IsPaused = false

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	if p.client.Self.Channel != nil {
		msg := fmt.Sprintf(`Reproduciendo archivo local: <b>%s</b> (Añadido por %s)`, track.Title, track.AddedBy)
		p.client.Self.Channel.Send(msg, false)
	}
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}

	return p.PlayDirectStream(ctx, track.URL)
}

// PlayDirectStream lanza ffmpeg y envía el audio a Mumble
func (p *Player) PlayDirectStream(ctx context.Context, streamURL string) error {
	exeFFmpeg := ResolveExecutable("ffmpeg")
	
	// Argumentos para FFmpeg: leer de la URL, formato s16le, 48000Hz, MONO (gumble solo soporta 1 canal)
	// Añadimos flags de reconexión para evitar que YouTube corte el stream a la mitad (Error -10054)
	cmdFFmpeg := exec.CommandContext(ctx, exeFFmpeg,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", streamURL,
		"-ac", "1",
		"-ar", "48000",
		"-f", "s16le",
		"pipe:1",
	)

	var errOut bytes.Buffer
	cmdFFmpeg.Stderr = &errOut

	stdout, err := cmdFFmpeg.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmdFFmpeg.Start(); err != nil {
		return fmt.Errorf("error iniciando ffmpeg: %v, stderr: %s", err, errOut.String())
	}

	outChan := p.client.AudioOutgoing()

	defer cmdFFmpeg.Wait()

	// Leer frames de 10ms (480 samples por canal, 1 canal = 480 totales = 960 bytes)
	// gumble usa 10ms por defecto y 1 canal (Mono).
	bufferSize := 480
	
	// Ticker exacto a 10ms
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		intBuf := make(gumble.AudioBuffer, bufferSize)
		
		select {
		case <-ctx.Done():
			return nil // Detenido manualmente
		case <-ticker.C:
			// Si está pausado, no leemos de ffmpeg ni escribimos a mumble
			if p.IsPaused {
				continue
			}

			// Leer binario directamente al slice de int16
			err := binary.Read(stdout, binary.LittleEndian, &intBuf)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					stderrStr := errOut.String()
					if stderrStr != "" {
						fmt.Printf("FFmpeg finalizó. Stderr: %s\n", stderrStr)
					} else {
						fmt.Println("Fin de la canción (EOF normal).")
					}
				} else {
					fmt.Printf("Error leyendo FFmpeg: %v\n", err)
				}
				return nil
			}

			// Actualizar posición (10ms transcurridos en tiempo real de audio, ajustado por velocidad)
			// Ya que leemos 10ms cada vez
			p.Position += time.Duration(10 * float32(time.Millisecond) * p.Speed)
			
			// Notificar cambio de estado cada ~1 segundo (100 iteraciones de 10ms)
			if int(p.Position.Milliseconds()/10)%100 == 0 {
				if p.OnStateChange != nil {
					p.OnStateChange()
				}
			}

			// Aplicar Ducking y Volumen Base
			vol := p.BaseVolume
			if p.duckingCount.Load() > 0 {
				vol *= p.DuckingVol
			}
			if vol != 1.0 {
				for i := range intBuf {
					intBuf[i] = int16(float32(intBuf[i]) * vol)
				}
			}

			// Escribir a Mumble
			outChan <- intBuf
		}
	}
}

func (p *Player) Pause() {
	p.IsPaused = true
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
}

func (p *Player) Resume() {
	p.IsPaused = false
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
}

func (p *Player) Stop() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.CurrentTrack = nil
	p.Position = 0
	p.IsPaused = false
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
}

func (p *Player) StartDucking() {
	p.duckingCount.Add(1)
}

func (p *Player) StopDucking() {
	p.duckingCount.Add(-1)
}
