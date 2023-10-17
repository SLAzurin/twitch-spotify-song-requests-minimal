package data

type ConfigDatabase struct {
	TwitchChannel            string `env:"TWITCH_CHANNEL"`
	TwitchAccount            string `env:"TWITCH_ACCOUNT"`
	TwitchPassword           string `env:"TWITCH_PASSWORD"`
	TwitchAPIClientID        string `env:"TWITCH_CLIENT_ID"`
	TwitchAPIClientSecret    string `env:"TWITCH_CLIENT_SECRET"`
	TwitchAPIRefreshToken    string `env:"TWITCH_REFRESH_TOKEN"`
	TwitchPermsOverrideUsers string `env:"TWITCH_PERMS_OVERRIDE_USERS"`
	TwitchSongReqestRewardID string `env:"TWITCH_SONG_REQUEST_REWARD_ID"`
	SpotifyID                string `env:"SPOTIFY_ID"`
	SpotifySecret            string `env:"SPOTIFY_SECRET"`
	SpotifyRegion            string `env:"SPOTIFY_REGION"`
}

var AppCfg ConfigDatabase
