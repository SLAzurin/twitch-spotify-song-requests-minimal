package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/nicklaw5/helix/v2"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/api"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/data"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/utils"
)

var helixMainClient *helix.Client

func main() {
	cleanenv.ReadEnv(&data.AppCfg)
	cleanenv.ReadConfig("channels.env", &data.AppCfg)
	cleanenv.ReadConfig(".env", &data.AppCfg)
	err := utils.ValidateConf(data.AppCfg)
	if err != nil {
		log.Fatalln(err)
	}

	api.StartupSpotify()

	helixMainClient, err = helix.NewClient(&helix.Options{
		ClientID:        data.AppCfg.TwitchAPIClientID,
		UserAccessToken: data.AppCfg.TwitchPassword,
	})
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for {
			ircConn, _ := api.RunIRC(bot, struct {
				Password        string
				Nickname        string
				Channel         string
				HelixMainClient *helix.Client
			}{
				Password:        data.AppCfg.TwitchPassword,
				Nickname:        data.AppCfg.TwitchAccount,
				Channel:         data.AppCfg.TwitchChannel,
				HelixMainClient: helixMainClient,
			})
			<-ircConn.ExitCh
			ircConn.Conn.Close()
		}
	}()

	select {}
}
