package mumble

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
	"mumblebeats/internal/audio"
	"mumblebeats/internal/config"
	"mumblebeats/internal/db"
)

type BotClient struct {
	Config *config.Config
	Client *gumble.Client
	Player *audio.Player
}

func NewBotClient(cfg *config.Config) *BotClient {
	return &BotClient{
		Config: cfg,
	}
}

func (b *BotClient) Connect() error {
	mumbleConfig := gumble.NewConfig()
	mumbleConfig.Username = b.Config.Username
	mumbleConfig.Password = b.Config.Password

	// Aumentar el tamaño del frame de audio para mejorar el bitrate y la calidad.
	// Por defecto es 40 (32 kbps). Al subirlo a 120, permitimos hasta 96 kbps.
	mumbleConfig.AudioDataBytes = 120

	// Manejador de eventos (Conexión, Desconexión, Mensajes de texto)
	mumbleConfig.Attach(gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("Conectado al servidor de Mumble: %s\n", b.Config.ServerAddress)
			
			// Inicializar reproductor de audio
			b.Player = audio.NewPlayer(e.Client)

			// Unirse al canal configurado
			if b.Config.Channel != "" {
				channel := e.Client.Channels.Find(b.Config.Channel)
				if channel != nil {
					e.Client.Self.Move(channel)
				}
			}

			// Saludar por texto
			message := "¡Hola! MumbleBeats conectado y listo para poner música. Escribe !help para ver los comandos."
			if e.Client.Self.Channel != nil {
				e.Client.Self.Channel.Send(message, false)
			}
		},
		Disconnect: func(e *gumble.DisconnectEvent) {
			fmt.Printf("Desconectado de Mumble: %v\n", e.Type)
			if e.Type == gumble.DisconnectError {
				fmt.Printf("Razón del error: %v\n", e.String)
			}
			os.Exit(1)
		},
		TextMessage: func(e *gumble.TextMessageEvent) {
			senderName := "Servidor"
			if e.Sender != nil {
				senderName = e.Sender.Name
			}
			fmt.Printf("Mensaje recibido de %s: %s\n", senderName, e.Message)

			// Solo procesar comandos de usuarios reales
			if e.Sender == nil {
				return
			}

			// Parsear comandos
			plainMessage := gumbleutil.PlainText(&e.TextMessage)
			plainMessage = strings.TrimSpace(plainMessage)

			if len(plainMessage) > 6 && strings.HasPrefix(plainMessage, "!play ") || len(plainMessage) > 3 && strings.HasPrefix(plainMessage, "!p ") {
				var query string
				if strings.HasPrefix(plainMessage, "!p ") {
					query = strings.TrimSpace(plainMessage[3:])
				} else {
					query = strings.TrimSpace(plainMessage[6:])
				}
				e.Sender.Send("Buscando y extrayendo información...")
				
				// Buscar metadata en una goroutine para no bloquear
				go func() {
					if strings.Contains(query, "list=") || strings.Contains(query, "playlist?") {
						tracks, err := audio.FetchPlaylist(query)
						if err != nil {
							e.Sender.Send(fmt.Sprintf("Error cargando playlist: %v", err))
							return
						}
						e.Sender.Send(fmt.Sprintf("Playlist encontrada con %d canciones. Añadiendo a la cola...", len(tracks)))
						for i, t := range tracks {
							id, err := db.AddTrack(t.Title, t.WebpageURL, "youtube", senderName, t.Thumbnail)
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

					metadata, err := audio.FetchMetadata(query)
					if err != nil {
						e.Sender.Send(fmt.Sprintf("Error buscando '%s': %v", query, err))
						return
					}

					id, err := db.AddTrack(metadata.Title, metadata.WebpageURL, "youtube", senderName, metadata.Thumbnail)
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
			} else if len(plainMessage) > 10 && strings.HasPrefix(plainMessage, "!playlist ") {
				name := strings.TrimSpace(plainMessage[10:])
				err := db.LoadPlaylist(name, senderName)
				if err != nil {
					e.Sender.Send(fmt.Sprintf("Error cargando playlist '%s': %v", name, err))
				} else {
					e.Sender.Send(fmt.Sprintf("Playlist '%s' añadida a la cola.", name))
				}
			} else if len(plainMessage) > 7 && strings.HasPrefix(plainMessage, "!radio ") {
				url := strings.TrimSpace(plainMessage[7:])
				id, err := db.AddTrack("Radio", url, "radio", senderName, "")
				if err != nil {
					e.Sender.Send(fmt.Sprintf("Error añadiendo radio: %v", err))
				} else {
					e.Sender.Send(fmt.Sprintf("Radio añadida a la cola (ID: %d)", id))
				}

			} else if len(plainMessage) > 11 && strings.HasPrefix(plainMessage, "!playlocal ") {
				nombre := strings.TrimSpace(plainMessage[11:])
				e.Sender.Send(fmt.Sprintf("Buscando '%s' en la biblioteca local...", nombre))
				
				// Buscar sincrónicamente (suele ser rápido)
				foundPath, foundName, err := db.FindLocalFile(nombre)
				if err != nil {
					e.Sender.Send(fmt.Sprintf("Error: %v", err))
				} else {
					id, err := db.AddTrack(foundName, foundPath, "local", senderName, "")
					if err != nil {
						e.Sender.Send(fmt.Sprintf("Error añadiendo a la cola: %v", err))
					} else {
						e.Sender.Send(fmt.Sprintf("'%s' añadido a la cola (ID: %d)", foundName, id))
					}
				}

			} else if plainMessage == "!skip" || plainMessage == "!s" {
				if !isAdmin(b.Config.Admins, senderName) {
					e.Sender.Send("❌ No tienes permisos de administrador para usar este comando.")
					return
				}
				if b.Player != nil {
					b.Player.Stop()
					if e.Client.Self.Channel != nil {
						e.Client.Self.Channel.Send(fmt.Sprintf("Canción saltada por %s.", senderName), false)
					}
				}
			} else if plainMessage == "!clear" {
				if !isAdmin(b.Config.Admins, senderName) {
					e.Sender.Send("❌ No tienes permisos de administrador para usar este comando.")
					return
				}
				err := db.ClearQueue()
				if err != nil {
					e.Sender.Send(fmt.Sprintf("Error limpiando la cola: %v", err))
				} else {
					if e.Client.Self.Channel != nil {
						e.Client.Self.Channel.Send(fmt.Sprintf("La cola ha sido limpiada por %s.", senderName), false)
					}
				}
			} else if plainMessage == "!queue" || plainMessage == "!q" {
				tracks, err := db.GetQueue(5)
				if err != nil {
					e.Sender.Send(fmt.Sprintf("Error obteniendo cola: %v", err))
					return
				}
				if len(tracks) == 0 {
					e.Sender.Send("La cola está vacía.")
				} else {
					msg := "<b>Siguientes canciones en cola:</b><br><br><table style=\"border-spacing: 0 5px;\">"
					for i, t := range tracks {
						imgTag := ""
						if t.Thumbnail != "" {
							imgBase64 := audio.GetThumbnailBase64(t.Thumbnail, "default")
							if imgBase64 != "" {
								imgTag = fmt.Sprintf(`<img src="%s" height="40" style="vertical-align: middle; border-radius: 4px;" />`, imgBase64)
							}
						}
						msg += fmt.Sprintf(`<tr><td style="padding-right: 10px;"><b>%d.</b></td><td style="padding-right: 10px;">%s</td><td>%s <span style="color: #888; font-size: 0.9em;">(por %s)</span></td></tr>`, i+1, imgTag, t.Title, t.AddedBy)
					}
					msg += "</table>"
					e.Sender.Send(msg)
				}
			} else if plainMessage == "!help" || plainMessage == "!h" {
				helpMsg := `<b>Comandos de MumbleBeats:</b><br>
<br>
<b>!play &lt;url/nombre&gt;</b> o <b>!p</b> - Añade una canción de YouTube (por enlace o búsqueda).<br>
<b>!playlist &lt;nombre&gt;</b> - Carga una playlist guardada en la base de datos.<br>
<b>!radio &lt;url&gt;</b> - Añade una transmisión de radio en directo.<br>
<b>!playlocal &lt;nombre&gt;</b> - Añade un archivo .mp3 de la carpeta local.<br>
<b>!queue</b> o <b>!q</b> - Muestra las siguientes 5 canciones en la cola.<br>
<b>!now</b> o <b>!np</b> - Muestra la canción actual y su progreso.<br>
<br>
🌐 <b>Panel Web:</b> <code>http://localhost:8080</code> (Para gestionar la cola y playlists).<br>
<br>
<b>[Admins] !pause / !resume</b> - Pausa o reanuda la canción actual.<br>
<b>[Admins] !skip</b> o <b>!s</b> - Salta la canción que está sonando actualmente.<br>
<b>[Admins] !clear</b> - Elimina todas las canciones de la cola.<br>
<b>[Admins] !stop</b> - Detiene la música sin limpiar la cola.<br>
<b>!help</b> o <b>!h</b> - Muestra este mensaje de ayuda.`
				e.Sender.Send(helpMsg)
			} else if plainMessage == "!now" || plainMessage == "!np" {
				if b.Player != nil && b.Player.CurrentTrack != nil {
					track := b.Player.CurrentTrack
					pos := int(b.Player.Position.Seconds())
					
					estado := "▶️ Reproduciendo"
					if b.Player.IsPaused {
						estado = "⏸️ Pausado"
					}
					
					imgTag := ""
					if track.Thumbnail != "" {
						imgBase64 := audio.GetThumbnailBase64(track.Thumbnail, "mqdefault")
						if imgBase64 != "" {
							imgTag = fmt.Sprintf(`<br/><img src="%s" height="90" />`, imgBase64)
						}
					}
					
					e.Sender.Send(fmt.Sprintf("%s ahora: <b>%s</b><br>Progreso: %02d:%02d%s", estado, track.Title, pos/60, pos%60, imgTag))
				} else {
					e.Sender.Send("No hay ninguna canción sonando ahora mismo.")
				}
			} else if plainMessage == "!pause" {
				if !isAdmin(b.Config.Admins, senderName) {
					e.Sender.Send("❌ No tienes permisos de administrador para usar este comando.")
					return
				}
				if b.Player != nil {
					b.Player.Pause()
					e.Sender.Send("⏸️ Reproducción pausada.")
				}
			} else if plainMessage == "!resume" {
				if !isAdmin(b.Config.Admins, senderName) {
					e.Sender.Send("❌ No tienes permisos de administrador para usar este comando.")
					return
				}
				if b.Player != nil {
					b.Player.Resume()
					e.Sender.Send("▶️ Reproducción reanudada.")
				}
			} else if plainMessage == "!stop" {
				if !isAdmin(b.Config.Admins, senderName) {
					e.Sender.Send("❌ No tienes permisos de administrador para usar este comando.")
					return
				}
				// Detiene el reproductor, pero NO limpia la cola (el Worker seguirá con el siguiente tras el sleep).
				// ¡Ojo! Como el worker intentará la siguiente si salimos de 'Stop()', 
				// en realidad !stop actuará casi como !skip si no pausamos el worker. 
				// Para detener de verdad, limpiamos la cola y luego paramos:
				db.ClearQueue()
				if b.Player != nil {
					b.Player.Stop()
					if e.Client.Self.Channel != nil {
						e.Client.Self.Channel.Send("Reproducción y cola detenidas.", false)
					}
				}
			}
		},
	})

	address := net.JoinHostPort(b.Config.ServerAddress, b.Config.ServerPort)
	
	tlsConfig := &tls.Config{
		InsecureSkipVerify: b.Config.Insecure, // Permitir certificados autofirmados
	}

	mumbleConfig.AttachAudio(&audioListener{bot: b})

	client, err := gumble.DialWithDialer(new(net.Dialer), address, mumbleConfig, tlsConfig)
	if err != nil {
		return err
	}

	b.Client = client
	return nil
}

// audioListener escucha cuando alguien habla en el canal de Mumble
type audioListener struct {
	bot *BotClient
}

func (l *audioListener) OnAudioStream(e *gumble.AudioStreamEvent) {
	if l.bot.Player != nil {
		l.bot.Player.StartDucking()
	}
	
	// Consumir paquetes de audio en una goroutine para no bloquear el cliente de gumble
	go func() {
		if l.bot.Player != nil {
			defer l.bot.Player.StopDucking()
		}
		for range e.C {
			// Descartamos los paquetes, solo nos importa el evento de voz
		}
	}()
}

func isAdmin(admins []string, username string) bool {
	for _, admin := range admins {
		if strings.EqualFold(admin, username) {
			return true
		}
	}
	return false
}
