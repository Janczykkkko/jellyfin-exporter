package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// GetSessions fetches sessions from Jellyfin
func GetSessions() {
	var (
		JellyJSON []JellySession
		count     int
	)
	url := jellyfinAddress + "/Sessions?api_key=" + jellyfinApiKey
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching sessions: " + err.Error())
	}
	defer resp.Body.Close()
	log.Printf("API request to %s completed with status code: %d", jellyfinAddress, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error fetching sessions: " + err.Error())
	}
	err = json.Unmarshal(body, &JellyJSON)
	if err != nil {
		fmt.Println("Error fetching sessions: " + err.Error())
	}
	count = 0
	for _, obj := range JellyJSON {
		if len(obj.NowPlayingQueueFullItems) > 0 &&
			obj.PlayState.PlayMethod != "" {
			var userName string
			var name string
			var state string
			var bitrate string
			var substream string
			var playMethod string
			var deviceName string
			if obj.PlayState.IsPaused {
				state = "paused"
			} else {
				state = "in progress"
			}
			if err == nil {
				bitrateFloat := float64(obj.NowPlayingQueueFullItems[0].MediaSources[0].Bitrate) / 1000000.0
				bitrate = strconv.FormatFloat(bitrateFloat, 'f', -1, 64)
			} else {
				bitrate = "error"
			}
			name = obj.NowPlayingQueueFullItems[0].MediaSources[0].Name
			userName = obj.UserName
			playMethod = obj.PlayState.PlayMethod
			deviceName = obj.DeviceName
			SubtitleStreamIndex := obj.PlayState.SubtitleStreamIndex
			if SubtitleStreamIndex >= 0 && SubtitleStreamIndex < len(obj.NowPlayingQueueFullItems[0].MediaStreams) {
				substream = obj.NowPlayingQueueFullItems[0].MediaStreams[obj.PlayState.SubtitleStreamIndex].DisplayTitle
			} else {
				substream = "None"
			}
			count++
			updateSessionMetrics(userName, state, name, playMethod, substream, deviceName, bitrate, count)
		} else {
			continue
		}

	}

}

func updateSessionMetrics(username, state, name, playMethod, substream, deviceName string, bitrate string, count int) {
	sessionLabels := prometheus.Labels{
		"UserName":   username,
		"State":      state,
		"Name":       name,
		"Bitrate":    bitrate,
		"PlayMethod": playMethod,
		"Substream":  substream,
		"DeviceName": deviceName,
	}

	// Set labels and update the gauge for the specific session
	sessionsMetric.With(sessionLabels).Set(float64(count))
}