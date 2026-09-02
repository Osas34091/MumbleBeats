package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func main() {
	query := "ytsearch1:avicii levels"
	cmdYt := exec.Command("yt-dlp", "--no-playlist", "-J", "-f", "bestaudio", query)
	var out bytes.Buffer
	cmdYt.Stdout = &out
	err := cmdYt.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	var rawData map[string]interface{}
	json.Unmarshal(out.Bytes(), &rawData)
	
	entries := rawData["entries"].([]interface{})
	firstEntry := entries[0].(map[string]interface{})
	
	fmt.Printf("thumbnail type: %T\n", firstEntry["thumbnail"])
	fmt.Printf("thumbnail value: %v\n", firstEntry["thumbnail"])
	
	fmt.Printf("title: %v\n", firstEntry["title"])
}
