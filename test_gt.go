package main

import (
	"fmt"
	"mumblebeats/internal/audio"
)

func main() {
	meta, err := audio.FetchMetadata("ytsearch:GT rap porta")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Thumbnail:", meta.Thumbnail)
}
