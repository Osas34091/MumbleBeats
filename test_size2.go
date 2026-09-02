package main

import (
	"bytes"
	"fmt"
	"image/gif"
	"image/jpeg"
	"net/http"
	"github.com/nfnt/resize"
)

func main() {
	resp, _ := http.Get("https://i.ytimg.com/vi/WFwYEe2REaA/hqdefault.jpg")
	defer resp.Body.Close()
	img, _ := jpeg.Decode(resp.Body)
	
	m := resize.Resize(0, 32, img, resize.Lanczos3)
	
	var buf bytes.Buffer
	gif.Encode(&buf, m, &gif.Options{NumColors: 8})
	
	fmt.Printf("32px GIF 8 colors size: %d bytes\n", buf.Len())
}
