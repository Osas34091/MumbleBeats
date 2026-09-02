package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"net/http"
	"github.com/nfnt/resize"
)

func main() {
	resp, _ := http.Get("https://i.ytimg.com/vi/WFwYEe2REaA/hqdefault.jpg")
	defer resp.Body.Close()
	img, _, _ := jpeg.Decode(resp.Body)
	
	m := resize.Resize(0, 32, img, resize.Lanczos3)
	
	var buf bytes.Buffer
	jpeg.Encode(&buf, m, &jpeg.Options{Quality: 40})
	
	fmt.Printf("32px Quality 40 size: %d bytes\n", buf.Len())
	
	m2 := resize.Resize(0, 40, img, resize.Lanczos3)
	var buf2 bytes.Buffer
	jpeg.Encode(&buf2, m2, &jpeg.Options{Quality: 50})
	fmt.Printf("40px Quality 50 size: %d bytes\n", buf2.Len())
}
