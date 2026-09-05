package i18n

import (
	"strings"
)

var translations = map[string]map[string]string{
	"en": {
		"connected": "Hello! MumbleBeats connected and ready to play music. Type !help to see commands.",
		"disconnected_mumble": "Disconnected from Mumble: %v",
		"error_reason": "Error reason: %v",
		"added_by": "Added by",
		"playing_radio": "Playing Radio",
		"playing_local": "Playing Local File",
		"playing_yt": "Playing",
		"cmd_help_title": "--- MumbleBeats Commands ---",
		"cmd_help_play": "!play <url|search> : Play or add to queue",
		"cmd_help_radio": "!radio <url>       : Play an internet radio stream",
		"cmd_help_local": "!local <file>      : Play a local file from the music folder",
		"cmd_help_skip": "!skip              : Skip the current track",
		"cmd_help_stop": "!stop              : Stop playback and clear the queue",
		"cmd_help_pause": "!pause             : Pause playback",
		"cmd_help_resume": "!resume            : Resume playback",
		"cmd_help_queue": "!queue             : Show the current queue",
		"cmd_help_volume": "!volume <0-100>    : Change the volume",
		"cmd_not_found": "Unknown command. Type !help for the list.",
		"queued": "Added to queue",
		"queue_empty": "The queue is empty.",
		"queue_title": "--- Current Queue ---",
		"queue_item": "%d. %s (Added by %s)",
		"skipped": "Skipped current track.",
		"stopped": "Playback stopped and queue cleared.",
		"paused": "Playback paused.",
		"resumed": "Playback resumed.",
		"volume_set": "Volume set to %d%%.",
		"volume_invalid": "Volume must be between 0 and 100.",
	},
	"es": {
		"connected": "¡Hola! MumbleBeats conectado y listo para poner música. Escribe !help para ver los comandos.",
		"disconnected_mumble": "Desconectado de Mumble: %v",
		"error_reason": "Razón del error: %v",
		"added_by": "Añadido por",
		"playing_radio": "Reproduciendo Radio",
		"playing_local": "Reproduciendo Archivo Local",
		"playing_yt": "Reproduciendo",
		"cmd_help_title": "--- Comandos de MumbleBeats ---",
		"cmd_help_play": "!play <url|busqueda> : Reproduce o añade a la cola",
		"cmd_help_radio": "!radio <url>         : Reproduce una radio por internet",
		"cmd_help_local": "!local <archivo>     : Reproduce un archivo de la carpeta music",
		"cmd_help_skip": "!skip                : Salta la canción actual",
		"cmd_help_stop": "!stop                : Detiene la música y limpia la cola",
		"cmd_help_pause": "!pause               : Pausa la música",
		"cmd_help_resume": "!resume              : Reanuda la música",
		"cmd_help_queue": "!queue               : Muestra la cola actual",
		"cmd_help_volume": "!volume <0-100>      : Cambia el volumen",
		"cmd_not_found": "Comando no reconocido. Escribe !help.",
		"queued": "Añadido a la cola",
		"queue_empty": "La cola está vacía.",
		"queue_title": "--- Cola Actual ---",
		"queue_item": "%d. %s (Añadido por %s)",
		"skipped": "Canción saltada.",
		"stopped": "Reproducción detenida y cola limpiada.",
		"paused": "Reproducción pausada.",
		"resumed": "Reproducción reanudada.",
		"volume_set": "Volumen ajustado al %d%%.",
		"volume_invalid": "El volumen debe estar entre 0 y 100.",
	},
}

// Get returns the translated string for a given key in the specified language.
// If the key or language doesn't exist, it falls back to English.
func Get(lang string, key string) string {
	lang = strings.ToLower(lang)
	if _, ok := translations[lang]; !ok {
		lang = "en"
	}
	
	if val, ok := translations[lang][key]; ok {
		return val
	}
	
	// Fallback to English if key missing
	if val, ok := translations["en"][key]; ok {
		return val
	}
	
	return key // Fallback to returning the key itself
}
