package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"mumblebeats/internal/db"
	"mumblebeats/internal/utils"

	"layeh.com/gumble/gumble"
)

type Player struct {
	client       *gumble.Client
	cancel       context.CancelFunc // Cancela la canción entera
	duckingCount atomic.Int32
	BaseVolume   float32
	DuckingVol   float32
	
	// Control de Reproducción
	CurrentTrack *db.Track
	StreamURL    string
	Position     time.Duration
	Speed        float32
	IsPaused     bool
	
	// Filtros y Seek
	Filter       string
	SeekOffset   int // en segundos
	restartChan  chan struct{} // Señal para reiniciar ffmpeg
	
	OnStateChange func() // Callback para notificar cambios (WebSockets)
}

func NewPlayer(client *gumble.Client) *Player {
	return &Player{
		client:      client,
		BaseVolume:  1.0,
		DuckingVol:  0.2, // 20% volume when ducking
		Speed:       1.0,
		restartChan: make(chan struct{}, 1),
	}
}

// PlayURL extrae la URL con yt-dlp y reproduce el audio
func (p *Player) PlayURL(track *db.Track) error {
	p.Stop() // Detener lo que esté sonando
	
	p.CurrentTrack = track
	p.IsPaused = false
	p.Position = 0
	p.SeekOffset = 0 // Resetear seek
	p.Filter = "off" // Resetear filtro
	p.Speed = 1.0    // Resetear velocidad
	p.IsPaused = false

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	fmt.Println("Extrayendo URL de audio con yt-dlp...")
	
	exeYtDlp := ResolveExecutable("yt-dlp")

	// Obtener la mejor URL de audio y metadata en formato JSON, ignorando playlists
	cmdYt := exec.CommandContext(ctx, exeYtDlp, "--no-playlist", "-J", "-f", "bestaudio", track.URL)
	utils.HideWindow(cmdYt) // Ocultar la ventana de consola en Windows
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

	for {
		// Drenar el canal de reinicio por si acaso
		select {
		case <-p.restartChan:
		default:
		}

		args := []string{
			"-reconnect", "1",
			"-reconnect_streamed", "1",
			"-reconnect_delay_max", "5",
		}
		
		if p.SeekOffset > 0 {
			args = append(args, "-ss", fmt.Sprintf("%d", p.SeekOffset))
		}
		
		args = append(args, "-i", streamURL)
		
		// Construir filtros de audio
		var filters []string
		
		if p.Speed != 1.0 {
			filters = append(filters, fmt.Sprintf("atempo=%.2f", p.Speed))
		}
		
		if p.Filter != "" && p.Filter != "off" {
			switch p.Filter {
			case "nightcore":
				filters = append(filters, "asetrate=48000*1.25,aresample=48000")
			case "bassboost":
				filters = append(filters, "bass=g=15:f=50:w=0.5")
			case "echo":
				filters = append(filters, "aecho=0.8:0.9:1000:0.3")
			}
		}
		
		if len(filters) > 0 {
			args = append(args, "-filter:a", strings.Join(filters, ","))
		}

		args = append(args, "-ac", "2", "-ar", "48000", "-f", "s16le", "pipe:1")

		cmdFFmpeg := exec.CommandContext(ctx, exeFFmpeg, args...)
		utils.HideWindow(cmdFFmpeg)

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

		// Leer frames de 10ms (480 samples por canal * 2 canales = 960)
		bufferSize := 960
		ticker := time.NewTicker(10 * time.Millisecond)
		
		restarting := false

		// Loop interno de lectura de FFmpeg
	readLoop:
		for {
			intBuf := make(gumble.AudioBuffer, bufferSize)
			
			select {
			case <-ctx.Done():
				ticker.Stop()
				cmdFFmpeg.Process.Kill()
				cmdFFmpeg.Wait()
				return nil // Detenido manualmente (siguiente canción o stop)
				
			case <-p.restartChan:
				// Petición de reinicio (Seek o Cambio de Filtro)
				restarting = true
				break readLoop
				
			case <-ticker.C:
				if p.IsPaused {
					continue
				}

				err := binary.Read(stdout, binary.LittleEndian, &intBuf)
				if err != nil {
					break readLoop // Fin del archivo o error (continuará al wait y saldrá)
				}

				p.Position += time.Duration(10 * float32(time.Millisecond) * p.Speed)
				
				if int(p.Position.Milliseconds()/10)%100 == 0 {
					if p.OnStateChange != nil {
						p.OnStateChange()
					}
				}

				vol := p.BaseVolume
				if p.duckingCount.Load() > 0 {
					vol *= p.DuckingVol
				}
				if vol != 1.0 {
					for i := range intBuf {
						intBuf[i] = int16(float32(intBuf[i]) * vol)
					}
				}

				outChan <- intBuf
			}
		}

		ticker.Stop()
		cmdFFmpeg.Process.Kill() // Matar el proceso anterior antes de reiniciar o salir
		cmdFFmpeg.Wait()
		
		if !restarting {
			// Si no estamos reiniciando, significa que la canción terminó
			return nil
		}
		// Si estamos reiniciando, el for exterior vuelve a lanzar FFmpeg con los nuevos parámetros
	}
}

// Métodos de Control PRO

func (p *Player) Seek(seconds int) error {
	p.SeekOffset = seconds
	p.Position = time.Duration(seconds) * time.Second
	
	// Notificar reinicio
	select {
	case p.restartChan <- struct{}{}:
	default:
	}
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
	return nil
}

func (p *Player) SetSpeed(speed float64) error {
	p.Speed = float32(speed)
	
	// Notificar reinicio
	select {
	case p.restartChan <- struct{}{}:
	default:
	}
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
	return nil
}

func (p *Player) ApplyFilter(filter string) error {
	p.Filter = filter
	
	// Notificar reinicio
	select {
	case p.restartChan <- struct{}{}:
	default:
	}
	
	if p.OnStateChange != nil {
		p.OnStateChange()
	}
	return nil
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
