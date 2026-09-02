package main

import (
	"fmt"
	"mumblebeats/internal/audio"
)

func main() {
	url := "https://i.ytimg.com/vi/44pt8w67S8I/hqdefault.jpg?sqp=-oaymwEcCNACELwBSFXyq4qpAw4IARUAAIhCGAFwAcABBg==&rs=AOn4CLBYNUpO61rEC69kiI3Z4QYi9cCe8A"
	b64 := audio.GetThumbnailBase64(url, "mqdefault")
	fmt.Printf("Base64 len: %d\n", len(b64))
	if len(b64) < 100 {
		fmt.Printf("Base64 string: %s\n", b64)
	}
}
