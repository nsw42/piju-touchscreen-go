package apiclient

import (
	"bytes"
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

var copyBuffer = make([]byte, 2048)

func extractIntFromJson(jsonObject map[string]any, key string, defaultVal int) int {
	tempFloat, ok := jsonObject[key].(float64)
	if !ok {
		return defaultVal
	} else {
		return int(tempFloat)
	}
}

func getStatusFromReply(reply map[string]any) Status {
	statStr := reply["PlayerStatus"]
	switch statStr {
	case "stopped":
		return Stopped
	case "playing":
		return Playing
	case "paused":
		return Paused
	default:
		// Error in reply
		fmt.Println("Unrecognised player status: ", statStr)
		return Stopped
	}
}

func getArtistTrackAndStream(reply map[string]any) (bool, string, string, string) {
	var currentTrack map[string]any
	isTrack := false
	streamName, _ := reply["CurrentStream"].(string)
	currentTrack, ok := reply["CurrentTrack"].(map[string]any)
	var artistName, trackName string
	if ok && len(currentTrack) > 0 {
		isTrack = true
		artistName, ok = currentTrack["artist"].(string)
		if !ok {
			artistName = "Unknown artist"
		}
		trackName, ok = currentTrack["title"].(string)
		if !ok {
			trackName = "Unknown track"
		}
	}
	return isTrack, artistName, trackName, streamName
}

func (client *Client) getArtworkFromReply(reply map[string]any) (string, []byte) {
	artworkUri, ok := reply["CurrentArtwork"].(string)
	if !ok {
		return "", nil
	}

	if artworkUri != client.CachedArtworkUri {
		// Need to update our cache
		client.CachedArtwork = client.fetchArtwork(artworkUri)
		if client.CachedArtwork != nil {
			// If the GET failed, we should be prepared to try again, so only update the
			// cached URI if the GET succeded
			client.CachedArtworkUri = artworkUri
		}
	}
	return artworkUri, client.CachedArtwork
}

func getScanningStatus(reply map[string]any) bool {
	workerStatusStr, ok := reply["WorkerStatus"].(string)
	if !ok {
		return false
	}
	return (strings.ToLower(workerStatusStr) != "idle")
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

	stat.Status = getStatusFromReply(reply)
	stat.IsTrack, stat.ArtistName, stat.TrackName, stat.StreamName = getArtistTrackAndStream(reply)
	stat.TrackNumber = extractIntFromJson(reply, "CurrentTrackIndex", 0)
	stat.AlbumTracks = extractIntFromJson(reply, "MaximumTrackIndex", 0)
	stat.ArtworkUri, stat.Artwork = client.getArtworkFromReply(reply)
	stat.Scanning = getScanningStatus(reply)

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
	var buffer bytes.Buffer
	_, err = io.CopyBuffer(&buffer, resp.Body, copyBuffer)
	if err != nil {
		return nil
	}
	return buffer.Bytes()
}

func (client *Client) SendPause() {
	client.SendSimpleCommand("player/pause", "pause")
}

func (client *Client) SendResume() {
	client.SendSimpleCommand("player/resume", "resume")
}

func (client *Client) SendResumeType(playerType string) {
	data := map[string]string{
		"player": playerType,
	}
	buf, _ := json.Marshal(data)
	body := bytes.NewReader(buf)
	resp, err := httpClient.Post(client.Host+"player/resume", "application/json", body)
	if err != nil {
		fmt.Println("Failed to send command to server: ", err)
		return
	}
	defer resp.Body.Close()
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
		fmt.Println("Failed to send "+operationDesc+" command to server: ", err)
		return
	}
	defer resp.Body.Close()
}
