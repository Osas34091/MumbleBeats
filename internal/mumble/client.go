package mumble

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
	"mumblebeats/internal/audio"
	"mumblebeats/internal/config"
)

type BotClient struct {
	Config       *config.Config
	Client       *gumble.Client
	Player       *audio.Player
	GlobalVolume float32
	OnStateChange func()
}

func NewBotClient(cfg *config.Config) *BotClient {
	return &BotClient{
		Config:       cfg,
		GlobalVolume: 1.0,
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
			if b.OnStateChange != nil {
				b.Player.OnStateChange = b.OnStateChange
			}

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
			b.Client = nil
			if b.Player != nil {
				b.Player.Stop()
			}
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

			if !strings.HasPrefix(plainMessage, "!") {
				return
			}

			// Parse command and args
			parts := strings.Split(plainMessage[1:], " ")
			cmdName := strings.ToLower(parts[0])
			args := parts[1:]

			cmd, exists := commandMap[cmdName]
			if !exists {
				e.Sender.Send(fmt.Sprintf("Comando no encontrado: !%s. Usa !help para ver la lista.", cmdName))
				return
			}

			if cmd.AdminOnly && !isAdmin(b.Config.Admins, senderName) {
				e.Sender.Send("❌ No tienes permisos de administrador para usar este comando.")
				return
			}

			cmd.Handler(b, e, args)
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

func (b *BotClient) Disconnect() {
	if b.Client != nil {
		b.Client.Disconnect()
		b.Client = nil
	}
	if b.Player != nil {
		b.Player.Stop()
	}
}

func (b *BotClient) Reconnect() error {
	b.Disconnect()
	fmt.Println("Intentando reconectar a Mumble...")
	return b.Connect()
}

type ChannelInfo struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

func (b *BotClient) GetChannels() []ChannelInfo {
	if b.Client == nil {
		return []ChannelInfo{}
	}
	var list []ChannelInfo
	for _, ch := range b.Client.Channels {
		list = append(list, ChannelInfo{
			ID:   ch.ID,
			Name: ch.Name,
		})
	}
	return list
}

func (b *BotClient) MoveToChannel(id uint32) error {
	if b.Client == nil {
		return fmt.Errorf("bot is not connected")
	}
	ch, exists := b.Client.Channels[id]
	if !exists {
		return fmt.Errorf("channel not found with ID: %d", id)
	}
	b.Client.Self.Move(ch)
	
	b.Config.Channel = ch.Name
	return config.SaveConfig(b.Config, "config.json")
}
