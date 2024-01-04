package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Host             string
	CachedArtworkUri string
	CachedArtwork    []byte
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func extractIntFromJson(jsonObject map[string]any, key string, defaultVal int) int {
	tempFloat, ok := jsonObject[key].(float64)
	if !ok {
		return defaultVal
	} else {
		return int(tempFloat)
	}
}

func (client *Client) GetCurrentStatus() NowPlaying {
	stat := NowPlaying{}
	stat.Status = Error // In case we return early
	resp, err := httpClient.Get(client.Host)
	if err != nil {
		fmt.Println("Error getting server status: ", err)
		return stat
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error getting server status: ", resp.StatusCode)
		return stat
	}

	var reply map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		fmt.Println("Error decoding JSON: ", err)
		return stat
	}

	statStr := reply["PlayerStatus"]
	var playerStatus Status
	if statStr == "stopped" {
		playerStatus = Stopped
	} else if statStr == "playing" {
		playerStatus = Playing
	} else if statStr == "paused" {
		playerStatus = Paused
	} else {
		// Error in reply
		fmt.Println("Unrecognised player status: ", statStr)
		return stat
	}

	var currentTrack map[string]any
	currentTrack, ok := reply["CurrentTrack"].(map[string]any)
	if !ok || len(currentTrack) == 0 {
		stat.IsTrack = false
		stat.StreamName, ok = reply["CurrentStream"].(string)
		if !ok {
			stat.StreamName = ""
		}
	} else {
		stat.IsTrack = true
		stat.ArtistName, ok = currentTrack["artist"].(string)
		if !ok {
			stat.ArtistName = "Unknown artist"
		}
		stat.TrackName, ok = currentTrack["title"].(string)
		if !ok {
			stat.TrackName = "Unknown track"
		}
	}
	stat.TrackNumber = extractIntFromJson(reply, "CurrentTrackIndex", 0)
	stat.AlbumTracks = extractIntFromJson(reply, "MaximumTrackIndex", 0)

	stat.ArtworkUri, ok = reply["CurrentArtwork"].(string)
	if ok {
		if stat.ArtworkUri != client.CachedArtworkUri {
			// Need to update our cache
			client.CachedArtwork = client.fetchArtwork(stat.ArtworkUri)
			if client.CachedArtwork != nil {
				// If the GET failed, we should be prepared to try again, so only update the
				// cached URI if the GET succeded
				client.CachedArtworkUri = stat.ArtworkUri
			}
		}
		stat.Artwork = client.CachedArtwork
	} else {
		fmt.Println("CurrentArtwork not a string: ", reply["CurrentArtwork"])
		stat.ArtworkUri = ""
		stat.Artwork = nil
	}

	stat.Status = playerStatus
	return stat
}

func (client *Client) fetchArtwork(uri string) []byte {
	if strings.HasPrefix(uri, "/") {
		uri = client.Host + uri
	}
	resp, err := httpClient.Get(uri)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	artwork, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return artwork
}

func (client *Client) SendPause() {
	client.SendSimpleCommand("player/pause", "pause")
}

func (client *Client) SendResume() {
	client.SendSimpleCommand("player/resume", "resume")
}

func (client *Client) SendNext() {
	client.SendSimpleCommand("player/next", "skip to next track")
}

func (client *Client) SendPrevious() {
	client.SendSimpleCommand("player/previous", "skip to previous track")
}

func (client *Client) SendSimpleCommand(uriSuffix string, operationDesc string) {
	resp, err := httpClient.Post(client.Host+uriSuffix, "application/json", nil)
	if err != nil {
		fmt.Println("Failed to send command to server: ", err)
		return
	}
	defer resp.Body.Close()
}
