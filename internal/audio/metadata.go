package audio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"mumblebeats/internal/utils"

	"github.com/nfnt/resize"
)

type TrackMetadata struct {
	Title      string `json:"title"`
	Uploader   string `json:"uploader"`
	Thumbnail  string `json:"thumbnail"`
	WebpageURL string `json:"webpage_url"`
	URL        string `json:"url"` // El stream de audio
}

// FetchMetadata obtiene la metadata de un video usando una URL o una búsqueda.
func FetchMetadata(query string) (*TrackMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exeYtDlp := ResolveExecutable("yt-dlp")

	// Si no es un enlace http, asumimos que es una búsqueda de YouTube
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		query = fmt.Sprintf("ytsearch1:%s", query)
	}

	cmdYt := exec.CommandContext(ctx, exeYtDlp, "--no-playlist", "-J", "-f", "bestaudio", query)
	utils.HideWindow(cmdYt)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmdYt.Stdout = &out
	cmdYt.Stderr = &errOut

	if err := cmdYt.Run(); err != nil {
		return nil, fmt.Errorf("error yt-dlp: %v, stderr: %s", err, errOut.String())
	}

	// yt-dlp con ytsearch1 devuelve a veces una lista de un elemento, vamos a parsear la salida cruda primero
	var rawData map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &rawData); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %v", err)
	}

	var metadata TrackMetadata

	// Si es un resultado de búsqueda, el formato JSON puede tener la key "entries"
	if entries, ok := rawData["entries"].([]interface{}); ok && len(entries) > 0 {
		if firstEntry, ok := entries[0].(map[string]interface{}); ok {
			entryBytes, _ := json.Marshal(firstEntry)
			json.Unmarshal(entryBytes, &metadata)
		} else {
			return nil, fmt.Errorf("no se encontraron entradas válidas en la búsqueda")
		}
	} else {
		// Si es una URL normal
		json.Unmarshal(out.Bytes(), &metadata)
	}

	// Fallback si WebpageURL está vacío
	if metadata.WebpageURL == "" && metadata.URL != "" {
		metadata.WebpageURL = metadata.URL
	} else if metadata.WebpageURL == "" {
		metadata.WebpageURL = query
	}
	
	// Si el thumbnail principal viene vacío pero hay un array de thumbnails, cogemos el mejor
	if metadata.Thumbnail == "" {
		var firstEntry map[string]interface{}
		if entries, ok := rawData["entries"].([]interface{}); ok && len(entries) > 0 {
			firstEntry, _ = entries[0].(map[string]interface{})
		} else {
			firstEntry = rawData
		}
		
		if firstEntry != nil {
			if thumbs, ok := firstEntry["thumbnails"].([]interface{}); ok && len(thumbs) > 0 {
				bestThumb, _ := thumbs[len(thumbs)-1].(map[string]interface{})
				if url, ok := bestThumb["url"].(string); ok {
					metadata.Thumbnail = url
				}
			}
		}
	}

	return &metadata, nil
}

// FetchPlaylist obtiene todas las canciones de una playlist de YouTube
func FetchPlaylist(url string) ([]*TrackMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	exeYtDlp := ResolveExecutable("yt-dlp")
	
	// --flat-playlist hace que no extraiga la información profunda (rápido)
	// --dump-json saca cada video en una línea JSON
	cmdYt := exec.CommandContext(ctx, exeYtDlp, "--flat-playlist", "-J", url)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmdYt.Stdout = &out
	cmdYt.Stderr = &errOut

	if err := cmdYt.Run(); err != nil {
		return nil, fmt.Errorf("error yt-dlp playlist: %v, stderr: %s", err, errOut.String())
	}

	var tracks []*TrackMetadata
	
	// Si escupe un solo JSON gigante con 'entries' (depende de la versión/flags)
	var rawData map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &rawData); err == nil {
		if entries, ok := rawData["entries"].([]interface{}); ok {
			for _, e := range entries {
				if entry, ok := e.(map[string]interface{}); ok {
					var meta TrackMetadata
					entryBytes, _ := json.Marshal(entry)
					json.Unmarshal(entryBytes, &meta)
					if meta.Thumbnail == "" {
						if thumbs, ok := entry["thumbnails"].([]interface{}); ok && len(thumbs) > 0 {
							bestThumb, _ := thumbs[len(thumbs)-1].(map[string]interface{})
							if url, ok := bestThumb["url"].(string); ok {
								meta.Thumbnail = url
							}
						}
					}
					if meta.Title != "" && meta.Title != "[Deleted video]" && meta.Title != "[Private video]" {
						if meta.WebpageURL == "" && meta.URL != "" {
							meta.WebpageURL = meta.URL
						}
						tracks = append(tracks, &meta)
					}
				}
			}
			return tracks, nil
		}
	}

	// Fallback por si escupe JSON lines
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var meta TrackMetadata
		if err := json.Unmarshal([]byte(line), &meta); err == nil {
			if meta.Thumbnail == "" {
				var raw map[string]interface{}
				if err := json.Unmarshal([]byte(line), &raw); err == nil {
					if thumbs, ok := raw["thumbnails"].([]interface{}); ok && len(thumbs) > 0 {
						bestThumb, _ := thumbs[len(thumbs)-1].(map[string]interface{})
						if url, ok := bestThumb["url"].(string); ok {
							meta.Thumbnail = url
						}
					}
				}
			}
			if meta.Title != "" && meta.Title != "[Deleted video]" && meta.Title != "[Private video]" {
				if meta.WebpageURL == "" && meta.URL != "" {
					meta.WebpageURL = meta.URL
				}
				tracks = append(tracks, &meta)
			}
		}
	}
	
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no se encontraron canciones válidas en la playlist")
	}

	return tracks, nil
}

// GetThumbnailBase64 descarga la miniatura y la devuelve en base64 para evitar bloqueos de Mumble.
// Comprime agresivamente para no superar el límite de 8KB de los mensajes de texto de Mumble.
func GetThumbnailBase64(url, sizeVariant string) string {
	if url == "" {
		return ""
	}
	
	thumb := url
	thumb = strings.ReplaceAll(thumb, ".webp", ".jpg")
	thumb = strings.ReplaceAll(thumb, "vi_webp", "vi")
	ytVariant := sizeVariant
	if sizeVariant == "now" {
		ytVariant = "mqdefault"
	}

	if ytVariant != "" {
		thumb = strings.ReplaceAll(thumb, "maxresdefault", ytVariant)
		thumb = strings.ReplaceAll(thumb, "sddefault", ytVariant)
		thumb = strings.ReplaceAll(thumb, "hqdefault", ytVariant)
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(thumb)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	imgData, _, err := image.Decode(resp.Body)
	if err != nil {
		return ""
	}

	var newHeight uint = 90
	quality := 40 // Calidad al 40% para evitar el límite de 5000 bytes en TODAS las imágenes (play, now, etc)
	
	if sizeVariant == "default" {
		newHeight = 90 // Mismo tamaño que play
		quality = 40
	} else if sizeVariant == "now" {
		quality = 40
	}
	
	m := resize.Resize(0, newHeight, imgData, resize.Lanczos3)
	
	// Intentamos comprimir, si el base64 final excede 4500 bytes (para no chocar con el límite de 5000 de Mumble),
	// reducimos la calidad dinámicamente.
	for q := quality; q >= 10; q -= 15 {
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, m, &jpeg.Options{Quality: q})
		if err != nil {
			return ""
		}
		
		b64 := fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
		if len(b64) < 4500 {
			return b64
		}
	}

	return ""
}
