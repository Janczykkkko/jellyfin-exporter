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
	sessionsMetric.Reset()
	count = 0
	for _, obj := range JellyJSON {
		if len(obj.NowPlayingQueueFullItems) > 0 &&
			obj.PlayState.PlayMethod != "" {
			var bitrate string
			var substream string
			bitrateFloat := float64(obj.NowPlayingQueueFullItems[0].MediaSources[0].Bitrate) / 1000000.0
			bitrate = strconv.FormatFloat(bitrateFloat, 'f', -1, 64)
			SubtitleStreamIndex := obj.PlayState.SubtitleStreamIndex
			if SubtitleStreamIndex >= 0 && SubtitleStreamIndex < len(obj.NowPlayingQueueFullItems[0].MediaStreams) {
				substream = obj.NowPlayingQueueFullItems[0].MediaStreams[obj.PlayState.SubtitleStreamIndex].DisplayTitle
			} else {
				substream = "None"
			}
			count = 1
			updateSessionMetrics(obj.UserName, obj.NowPlayingQueueFullItems[0].MediaSources[0].Name, obj.PlayState.PlayMethod, substream, obj.DeviceName, bitrate, count)
		} else if len(obj.FullNowPlayingItem.Container) > 0 &&
			obj.NowPlayingItem.Name != "" &&
			!obj.PlayState.IsPaused {
			var substream string = ""
			var bitrateData int
			for _, stream := range obj.NowPlayingItem.MediaStreams {
				if stream.Type == "Video" {
					bitrateData = stream.BitRate
				} else {
					bitrateData = 0
				}
			}
			bitrateFloat := float64(bitrateData) / 1000000.0
			bitrate := strconv.FormatFloat(bitrateFloat, 'f', -1, 64)
			count = 1
			updateSessionMetrics(obj.UserName, obj.NowPlayingItem.Name, obj.PlayState.PlayMethod, substream, obj.DeviceName, bitrate, count)
		} else {
			continue
		}
	}
}

func updateSessionMetrics(username, name, playMethod, substream, deviceName string, bitrate string, count int) {
	sessionLabels := prometheus.Labels{
		"UserName":   username,
		"Name":       name,
		"Bitrate":    bitrate,
		"PlayMethod": playMethod,
		"Substream":  substream,
		"DeviceName": deviceName,
	}

	// Set labels and update the gauge for the specific session
	sessionsMetric.With(sessionLabels).Set(float64(count))
}
