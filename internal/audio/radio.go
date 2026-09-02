package audio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type RadioStation struct {
	StationUUID string `json:"stationuuid"`
	Name        string `json:"name"`
	URL         string `json:"url_resolved"`
	Homepage    string `json:"homepage"`
	Favicon     string `json:"favicon"`
	Country     string `json:"country"`
	Language    string `json:"language"`
}

// SearchRadio busca una emisora por nombre en radio-browser.info
func SearchRadio(query string) ([]RadioStation, error) {
	// Endpoint de radio-browser
	apiURL := "https://de1.api.radio-browser.info/json/stations/search"
	
	reqURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	
	q := reqURL.Query()
	q.Set("name", query)
	q.Set("limit", "5")
	q.Set("hidebroken", "true")
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	
	// Radio-browser requires a custom User-Agent
	req.Header.Set("User-Agent", "MumbleBeats/1.0")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API de radio devolvió código %d", resp.StatusCode)
	}

	var stations []RadioStation
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, err
	}

	return stations, nil
}
