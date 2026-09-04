package mumble

import (
	"fmt"
	"strings"
	"time"

	"mumblebeats/internal/audio"
	"mumblebeats/internal/db"

	"layeh.com/gumble/gumble"
)

type CommandHandler func(b *BotClient, e *gumble.TextMessageEvent, args []string)

type Command struct {
	Name        string
	Aliases     []string
	Description string
	AdminOnly   bool
	Handler     CommandHandler
}

var BotCommands []Command

func init() {
	BotCommands = []Command{
	{
		Name:        "play",
		Aliases:     []string{"p"},
		Description: "Añade una canción de YouTube (por enlace o búsqueda).",
		Handler:     cmdPlay,
	},
	{
		Name:        "playlist",
		Description: "Carga una playlist guardada en la base de datos.",
		Handler:     cmdPlaylist,
	},
	{
		Name:        "radio",
		Description: "Añade una transmisión de radio en directo.",
		Handler:     cmdRadio,
	},
	{
		Name:        "playlocal",
		Description: "Añade un archivo .mp3 de la carpeta local.",
		Handler:     cmdPlayLocal,
	},
	{
		Name:        "queue",
		Aliases:     []string{"q"},
		Description: "Muestra las siguientes 5 canciones en la cola.",
		Handler:     cmdQueue,
	},
	{
		Name:        "now",
		Aliases:     []string{"np"},
		Description: "Muestra la canción actual y su progreso.",
		Handler:     cmdNow,
	},
	{
		Name:        "pause",
		Description: "Pausa la canción actual.",
		AdminOnly:   true,
		Handler:     cmdPause,
	},
	{
		Name:        "resume",
		Description: "Reanuda la canción actual.",
		AdminOnly:   true,
		Handler:     cmdResume,
	},
	{
		Name:        "skip",
		Aliases:     []string{"s"},
		Description: "Salta la canción que está sonando actualmente.",
		AdminOnly:   true,
		Handler:     cmdSkip,
	},
	{
		Name:        "clear",
		Description: "Elimina todas las canciones de la cola.",
		AdminOnly:   true,
		Handler:     cmdClear,
	},
	{
		Name:        "stop",
		Description: "Detiene la música y limpia la cola.",
		AdminOnly:   true,
		Handler:     cmdStop,
	},
	{
		Name:        "filter",
		Description: "Aplica un filtro DSP (nightcore, bassboost, echo, off).",
		AdminOnly:   true,
		Handler:     cmdFilter,
	},
	{
		Name:        "seek",
		Description: "Salta a un segundo específico de la canción (ej: !seek 60).",
		AdminOnly:   true,
		Handler:     cmdSeek,
	},
	{
		Name:        "speed",
		Description: "Cambia la velocidad (ej: !speed 1.5).",
		AdminOnly:   true,
		Handler:     cmdSpeed,
	},
		{
			Name:        "play",
			Aliases:     []string{"p"},
			Description: "Añade una canción de YouTube (por enlace o búsqueda).",
			Handler:     cmdPlay,
		},
		{
			Name:        "playlist",
			Description: "Carga una playlist guardada en la base de datos.",
			Handler:     cmdPlaylist,
		},
		{
			Name:        "radio",
			Description: "Añade una transmisión de radio en directo.",
			Handler:     cmdRadio,
		},
		{
			Name:        "playlocal",
			Description: "Añade un archivo .mp3 de la carpeta local.",
			Handler:     cmdPlayLocal,
		},
		{
			Name:        "queue",
			Aliases:     []string{"q"},
			Description: "Muestra las siguientes 5 canciones en la cola.",
			Handler:     cmdQueue,
		},
		{
			Name:        "now",
			Aliases:     []string{"np"},
			Description: "Muestra la canción actual y su progreso.",
			Handler:     cmdNow,
		},
		{
			Name:        "pause",
			Description: "Pausa la canción actual.",
			AdminOnly:   true,
			Handler:     cmdPause,
		},
		{
			Name:        "resume",
			Description: "Reanuda la canción actual.",
			AdminOnly:   true,
			Handler:     cmdResume,
		},
		{
			Name:        "skip",
			Aliases:     []string{"s"},
			Description: "Salta la canción que está sonando actualmente.",
			AdminOnly:   true,
			Handler:     cmdSkip,
		},
		{
			Name:        "clear",
			Description: "Elimina todas las canciones de la cola.",
			AdminOnly:   true,
			Handler:     cmdClear,
		},
		{
			Name:        "stop",
			Description: "Detiene la música y limpia la cola.",
			AdminOnly:   true,
			Handler:     cmdStop,
		},
		{
			Name:        "filter",
			Description: "Aplica un filtro DSP (nightcore, bassboost, echo, off).",
			AdminOnly:   true,
			Handler:     cmdFilter,
		},
		{
			Name:        "seek",
			Description: "Salta a un segundo específico de la canción (ej: !seek 60).",
			AdminOnly:   true,
			Handler:     cmdSeek,
		},
		{
			Name:        "speed",
			Description: "Cambia la velocidad (ej: !speed 1.5).",
			AdminOnly:   true,
			Handler:     cmdSpeed,
		},
		{
			Name:        "volume",
			Aliases:     []string{"v", "vol"},
			Description: "Ajusta el volumen (0 a 100).",
			AdminOnly:   true,
			Handler:     cmdVolume,
		},
		{
			Name:        "help",
			Aliases:     []string{"h"},
			Description: "Muestra este mensaje de ayuda.",
			Handler:     cmdHelp,
		},
	}
}

var commandMap = make(map[string]*Command)

func init() {
	for i := range BotCommands {
		cmd := &BotCommands[i]
		commandMap[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd
		}
	}
}

func getSenderName(e *gumble.TextMessageEvent) string {
	if e.Sender != nil {
		return e.Sender.Name
	}
	return "Servidor"
}

// Implementación de Comandos

func cmdHelp(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	var sb strings.Builder
	sb.WriteString("<b>Comandos de MumbleBeats:</b><br><ul>")
	for _, cmd := range BotCommands {
		aliasStr := ""
		if len(cmd.Aliases) > 0 {
			aliasStr = fmt.Sprintf(" (o !%s)", strings.Join(cmd.Aliases, ", !"))
		}
		adminStr := ""
		if cmd.AdminOnly {
			adminStr = " <span style='color:red;'>[Admin]</span>"
		}
		sb.WriteString(fmt.Sprintf("<li><b>!%s</b>%s%s - %s</li>", cmd.Name, aliasStr, adminStr, cmd.Description))
	}
	sb.WriteString("</ul>")
	e.Sender.Send(sb.String())
}

func cmdPlay(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if len(args) == 0 {
		e.Sender.Send("Uso: !play <búsqueda o enlace>")
		return
	}
	query := strings.Join(args, " ")
	senderName := getSenderName(e)
	e.Sender.Send("Buscando y extrayendo información...")

	go func() {
		if strings.Contains(query, "list=") || strings.Contains(query, "playlist?") {
			tracks, err := audio.FetchPlaylist(query)
			if err != nil {
				e.Sender.Send(fmt.Sprintf("Error cargando playlist: %v", err))
				return
			}
			e.Sender.Send(fmt.Sprintf("Playlist encontrada con %d canciones. Añadiendo a la cola...", len(tracks)))
			for i, t := range tracks {
				id, err := db.AddTrack(t.Title, t.WebpageURL, "youtube", senderName, t.Thumbnail, t.Duration)
				if i == 0 {
					if err == nil {
						imgTag := ""
						if t.Thumbnail != "" {
							imgBase64 := audio.GetThumbnailBase64(t.Thumbnail, "mqdefault")
							if imgBase64 != "" {
								imgTag = fmt.Sprintf(`<br/><img src="%s" height="90" />`, imgBase64)
							}
						}
						e.Sender.Send(fmt.Sprintf("Primera canción añadida a la cola (ID: %d)%s", id, imgTag))
					}
				}
			}
			e.Sender.Send("Todas las canciones de la playlist han sido añadidas.")
			return
		}

		isURL := strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://")
		if !isURL {
			foundPath, foundName, errLocal := db.FindLocalFile(query)
			if errLocal == nil {
				id, err := db.AddTrack(foundName, foundPath, "local", senderName, "", 0)
				if err != nil {
					e.Sender.Send(fmt.Sprintf("Error añadiendo a la cola: %v", err))
				} else {
					e.Sender.Send(fmt.Sprintf("'%s' añadido a la cola (ID: %d)", foundName, id))
				}
				return
			}
		}

		metadata, err := audio.FetchMetadata(query)
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error buscando '%s': %v", query, err))
			return
		}

		id, err := db.AddTrack(metadata.Title, metadata.WebpageURL, "youtube", senderName, metadata.Thumbnail, metadata.Duration)
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error añadiendo a la cola: %v", err))
		} else {
			imgTag := ""
			if metadata.Thumbnail != "" {
				imgBase64 := audio.GetThumbnailBase64(metadata.Thumbnail, "mqdefault")
				if imgBase64 != "" {
					imgTag = fmt.Sprintf(`<br/><img src="%s" height="90" />`, imgBase64)
				}
			}
			e.Sender.Send(fmt.Sprintf("'%s' añadido a la cola (ID: %d)%s", metadata.Title, id, imgTag))
		}
	}()
}

func cmdPlaylist(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if len(args) == 0 {
		e.Sender.Send("Uso: !playlist <nombre>")
		return
	}
	name := strings.Join(args, " ")
	err := db.LoadPlaylist(name, getSenderName(e))
	if err != nil {
		e.Sender.Send(fmt.Sprintf("Error cargando playlist '%s': %v", name, err))
	} else {
		e.Sender.Send(fmt.Sprintf("Playlist '%s' añadida a la cola.", name))
	}
}

func cmdRadio(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if len(args) == 0 {
		e.Sender.Send("Uso: !radio <url> o !radio <nombre>")
		return
	}
	query := strings.Join(args, " ")
	
	// Si es una URL, agregarla directamente
	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") {
		id, err := db.AddTrack("Radio", query, "radio", getSenderName(e), "", 0)
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error añadiendo radio: %v", err))
		} else {
			e.Sender.Send(fmt.Sprintf("Radio añadida a la cola (ID: %d)", id))
		}
		return
	}
	
	// Si no es URL, buscar emisora
	e.Sender.Send(fmt.Sprintf("Buscando emisora '%s'...", query))
	go func() {
		// Import "mumblebeats/internal/api" is required
		stations, err := audio.SearchRadio(query)
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error buscando emisora: %v", err))
			return
		}
		
		if len(stations) == 0 {
			e.Sender.Send("No se encontró ninguna emisora con ese nombre.")
			return
		}
		
		// Usar el primer resultado
		station := stations[0]
		id, err := db.AddTrack("Radio: "+station.Name, station.URL, "radio", getSenderName(e), station.Favicon, 0)
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error añadiendo emisora: %v", err))
		} else {
			e.Sender.Send(fmt.Sprintf("Emisora '%s' añadida a la cola (ID: %d)", station.Name, id))
		}
	}()
}

func cmdPlayLocal(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if len(args) == 0 {
		e.Sender.Send("Uso: !playlocal <nombre>")
		return
	}
	nombre := strings.Join(args, " ")
	e.Sender.Send(fmt.Sprintf("Buscando '%s' en la biblioteca local...", nombre))

	foundPath, foundName, err := db.FindLocalFile(nombre)
	if err != nil {
		e.Sender.Send(fmt.Sprintf("Error: %v", err))
	} else {
		id, err := db.AddTrack(foundName, foundPath, "local", getSenderName(e), "", 0)
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error añadiendo a la cola: %v", err))
		} else {
			e.Sender.Send(fmt.Sprintf("'%s' añadido a la cola (ID: %d)", foundName, id))
		}
	}
}

func cmdQueue(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	go func() {
		tracks, err := db.GetQueue(6) // Fetch 6 in case the first one is playing
		if err != nil {
			e.Sender.Send(fmt.Sprintf("Error obteniendo cola: %v", err))
			return
		}
		
		var pending []*db.Track
		for _, t := range tracks {
			if t.Status == "pending" {
				pending = append(pending, t)
			}
		}
		
		if len(pending) > 5 {
			pending = pending[:5]
		}

		if len(pending) == 0 {
			e.Sender.Send("La cola está vacía.")
		} else {
			for i, t := range pending {
				imgTag := ""
				if t.Thumbnail != "" {
					imgBase64 := audio.GetThumbnailBase64(t.Thumbnail, "default")
					if imgBase64 != "" {
						imgTag = fmt.Sprintf(`<br><img src="%s" height="90" style="vertical-align: middle; border-radius: 4px;" /><br>`, imgBase64)
					}
				}

				msg := ""
				if i == 0 {
					msg += "<b>Siguientes canciones en cola:</b><br><br>"
				}

				msg += fmt.Sprintf(`<b>%d.</b> %s <span style="color: #888; font-size: 0.9em;">(por %s)</span>%s`, i+1, t.Title, t.AddedBy, imgTag)
				e.Sender.Send(msg)
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
}

func cmdNow(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	go func() {
		if b.Player != nil && b.Player.CurrentTrack != nil {
			track := b.Player.CurrentTrack
			pos := int(b.Player.Position.Seconds())

			estado := "▶️ Reproduciendo"
			if b.Player.IsPaused {
				estado = "⏸️ Pausado"
			}

			imgTag := ""
			if track.Thumbnail != "" {
				imgBase64 := audio.GetThumbnailBase64(track.Thumbnail, "now")
				if imgBase64 != "" {
					imgTag = fmt.Sprintf(`<br/><img src="%s" height="90" />`, imgBase64)
				}
			}

			e.Sender.Send(fmt.Sprintf("%s ahora: <b>%s</b><br>Progreso: %02d:%02d%s", estado, track.Title, pos/60, pos%60, imgTag))
		} else {
			e.Sender.Send("No hay ninguna canción sonando ahora mismo.")
		}
	}()
}

func cmdPause(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if b.Player != nil {
		b.Player.Pause()
		e.Sender.Send("⏸️ Reproducción pausada.")
	}
}

func cmdResume(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if b.Player != nil {
		b.Player.Resume()
		e.Sender.Send("▶️ Reproducción reanudada.")
	}
}

func cmdSkip(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if b.Player != nil {
		b.Player.Stop()
		if e.Client.Self.Channel != nil {
			e.Client.Self.Channel.Send(fmt.Sprintf("Canción saltada por %s.", getSenderName(e)), false)
		}
	}
}

func cmdClear(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	err := db.ClearQueue()
	if err != nil {
		e.Sender.Send(fmt.Sprintf("Error limpiando la cola: %v", err))
	} else {
		if e.Client.Self.Channel != nil {
			e.Client.Self.Channel.Send(fmt.Sprintf("La cola ha sido limpiada por %s.", getSenderName(e)), false)
		}
	}
}

func cmdStop(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	db.ClearQueue()
	if b.Player != nil {
		b.Player.Stop()
		if e.Client.Self.Channel != nil {
			e.Client.Self.Channel.Send("Reproducción y cola detenidas.", false)
		}
	}
}

// Nuevos Comandos Nivel PRO
func cmdFilter(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if b.Player == nil || b.Player.CurrentTrack == nil {
		e.Sender.Send("No hay música reproduciéndose.")
		return
	}
	if len(args) == 0 {
		e.Sender.Send("Uso: !filter <nightcore|bassboost|echo|off>")
		return
	}
	
	filterName := strings.ToLower(args[0])
	e.Sender.Send(fmt.Sprintf("Aplicando filtro: %s...", filterName))
	err := b.Player.ApplyFilter(filterName)
	if err != nil {
		e.Sender.Send(fmt.Sprintf("Error aplicando filtro: %v", err))
	}
}

func cmdSeek(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if b.Player == nil || b.Player.CurrentTrack == nil {
		e.Sender.Send("No hay música reproduciéndose.")
		return
	}
	if len(args) == 0 {
		e.Sender.Send("Uso: !seek <segundos>")
		return
	}
	
	var secs int
	fmt.Sscanf(args[0], "%d", &secs)
	
	e.Sender.Send(fmt.Sprintf("Saltando a %d segundos...", secs))
	err := b.Player.Seek(secs)
	if err != nil {
		e.Sender.Send(fmt.Sprintf("Error buscando en la pista: %v", err))
	}
}

func cmdSpeed(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if b.Player == nil || b.Player.CurrentTrack == nil {
		e.Sender.Send("No hay música reproduciéndose.")
		return
	}
	if len(args) == 0 {
		e.Sender.Send("Uso: !speed <0.5 - 2.0>")
		return
	}
	
	var speed float64
	fmt.Sscanf(args[0], "%f", &speed)
	
	if speed < 0.5 || speed > 2.0 {
		e.Sender.Send("La velocidad debe estar entre 0.5 y 2.0")
		return
	}
	
	e.Sender.Send(fmt.Sprintf("Ajustando velocidad a %.2fx...", speed))
	err := b.Player.SetSpeed(speed)
	if err != nil {
		e.Sender.Send(fmt.Sprintf("Error ajustando velocidad: %v", err))
	}
}

func cmdVolume(b *BotClient, e *gumble.TextMessageEvent, args []string) {
	if len(args) < 1 {
		e.Sender.Send(fmt.Sprintf("Volumen actual: <b>%.0f%%</b>. Usa !volume <0-200>", b.GlobalVolume*100))
		return
	}
	var vol int
	_, err := fmt.Sscanf(args[0], "%d", &vol)
	if err != nil || vol < 0 || vol > 200 {
		e.Sender.Send("⚠️ Usa un número válido entre 0 y 200 para el volumen.")
		return
	}

	newVol := float32(vol) / 100.0
	b.GlobalVolume = newVol
	if b.Player != nil {
		b.Player.BaseVolume = newVol
	}
	e.Sender.Send(fmt.Sprintf("🔊 Volumen ajustado al <b>%d%%</b>", vol))
}
