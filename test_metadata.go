package main

import (
	"fmt"
	"mumblebeats/internal/audio"
)

func main() {
	meta, err := audio.FetchMetadata("ytsearch1:avicii levels")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Title: %s\n", meta.Title)
	fmt.Printf("Thumbnail: %s\n", meta.Thumbnail)
	
	// Test download base64
	imgBase64 := audio.GetThumbnailBase64(meta.Thumbnail, "mqdefault")
	fmt.Printf("mqdefault Base64 length: %d\n", len(imgBase64))
	
	imgBase64Def := audio.GetThumbnailBase64(meta.Thumbnail, "default")
	fmt.Printf("default Base64 length: %d\n", len(imgBase64Def))
}
