package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigDatabase struct {
	SpotifyID     string `env:"SPOTIFY_ID"`
	SpotifySecret string `env:"SPOTIFY_SECRET"`
}

func main() {
	appCfg := ConfigDatabase{}
	cleanenv.ReadEnv(&appCfg)
	cleanenv.ReadConfig(".env", &appCfg)
	client := &http.Client{}
	spotifyCredentials := base64.StdEncoding.EncodeToString([]byte(appCfg.SpotifyID + ":" + appCfg.SpotifySecret))
	refreshEndpoint := "https://accounts.spotify.com/api/token"
	body := url.Values{}
	body.Add("grant_type", "refresh_token")
	body.Add("refresh_token", os.Getenv("REFRESH_TOKEN"))
	BodyData := body.Encode()

	req, err := http.NewRequest("POST", refreshEndpoint, strings.NewReader(BodyData))
	if err != nil {
		log.Fatalln("New request error", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+spotifyCredentials)
	response, err := client.Do(req)
	if err != nil {
		log.Fatalln("Client do error", err)
	}
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln("Response body error", err)
	}
	spotifyToken := map[string]any{}
	err = json.Unmarshal(respBody, &spotifyToken)
	if err != nil {
		log.Fatalln("Body unmarshal error", err)
	}
	fmt.Println(spotifyToken["access_token"])
	defer response.Body.Close()
}
