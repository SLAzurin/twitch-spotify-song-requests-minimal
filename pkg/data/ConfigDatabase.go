package data

type ConfigDatabase struct {
	TwitchChannel            string `env:"TWITCH_CHANNEL"`
	TwitchAccount            string `env:"TWITCH_ACCOUNT"`
	TwitchPassword           string `env:"TWITCH_PASSWORD"`
	TwitchAPIClientID        string `env:"TWITCH_CLIENT_ID"`
	TwitchSongReqestRewardID string `env:"TWITCH_SONG_REQUEST_REWARD_ID"`
	SpotifyID                string `env:"SPOTIFY_ID"`
	SpotifySecret            string `env:"SPOTIFY_SECRET"`
}

var AppCfg ConfigDatabase
