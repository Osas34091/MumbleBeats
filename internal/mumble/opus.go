package mumble

import (
	"fmt"
	
	"github.com/kazzmir/opus-go/opus"
	"layeh.com/gumble/gumble"
)

func init() {
	gumble.RegisterAudioCodec(4, &OpusCodec{})
}

// OpusCodec implementa gumble.AudioCodec
type OpusCodec struct{}

func (c *OpusCodec) ID() int {
	return 4 // Opus ID en gumble
}

func (c *OpusCodec) NewEncoder() gumble.AudioEncoder {
	enc, err := NewOpusEncoder()
	if err != nil {
		return nil
	}
	return enc
}

func (c *OpusCodec) NewDecoder() gumble.AudioDecoder {
	return &DummyDecoder{} // Gumble crashea si devolvemos nil cuando alguien habla en el canal
}

type DummyDecoder struct{}

func (d *DummyDecoder) ID() int {
	return 4 // Opus
}

func (d *DummyDecoder) Decode(packet []byte, frameSize int) ([]int16, error) {
	// Devolvemos un buffer vacío porque no nos interesa decodificar/escuchar a los usuarios
	return make([]int16, frameSize), nil
}

func (d *DummyDecoder) Reset() {}

// OpusEncoder implementa gumble.AudioEncoder usando la librería pura en Go kazzmir/opus-go
type OpusEncoder struct {
	encoder *opus.Encoder
}

// NewOpusEncoder crea un nuevo codificador Opus puro en Go
func NewOpusEncoder() (*OpusEncoder, error) {
	// Sample rate: 48000Hz, Canales: 1 (Mono - requerido por gumble), Aplicación: Audio
	enc, err := opus.NewEncoder(48000, 1, opus.ApplicationAudio)
	if err != nil {
		return nil, err
	}
	
	return &OpusEncoder{
		encoder: enc,
	}, nil
}

// ID devuelve el identificador del codec Opus (4) según Mumble
func (e *OpusEncoder) ID() int {
	return 4 // 4 es el Opus ID en gumble
}

// Encode convierte PCM a Opus
func (e *OpusEncoder) Encode(pcm []int16, frameSize, maxDataBytes int) ([]byte, error) {
	packet := make([]byte, maxDataBytes)
	n, err := e.encoder.Encode(pcm, frameSize, packet)
	if err != nil {
		fmt.Printf("Error de codificación Opus: %v (frameSize=%d)\n", err, frameSize)
		return nil, err
	}
	return packet[:n], nil
}

// Reset reinicia el estado interno del encoder
func (e *OpusEncoder) Reset() {
	// kazzmir/opus-go/opus no provee una función Reset pública simple sin CGo
	// En la práctica, reconstruir el encoder o simplemente ignorar el reset suele funcionar bien para bots
	// ya que el flujo de audio es continuo y manejado por yt-dlp/ffmpeg.
}
