package utils

import (
	"log"
	"net/http"
	"strings"

	"github.com/nicklaw5/helix/v2"
)


func ChannelIsLive(client *helix.Client, channel string) (bool, error) {
	channel = strings.TrimPrefix(channel, "#")

	streams, err := client.GetStreams(&helix.StreamsParams{
		UserLogins: []string{channel},
	})
	if err != nil || streams.StatusCode != http.StatusOK {
		log.Println("Failed to check if broadcaster is live", err)
		return false, err
	}

	if len(streams.Data.Streams) > 0 {
		return true, nil
	}
	return false, nil
}
