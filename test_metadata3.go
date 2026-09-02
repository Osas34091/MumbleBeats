package main

import (
	"fmt"
	"mumblebeats/internal/audio"
)

func main() {
	meta, err := audio.FetchMetadata("avicii levels") // without prefix so it gets prefixed once
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Title: %s\n", meta.Title)
	fmt.Printf("Thumbnail: %s\n", meta.Thumbnail)
	
	b64 := audio.GetThumbnailBase64(meta.Thumbnail, "mqdefault")
	fmt.Printf("B64 length mqdefault: %d\n", len(b64))

	b64def := audio.GetThumbnailBase64(meta.Thumbnail, "default")
	fmt.Printf("B64 length default: %d\n", len(b64def))
}
