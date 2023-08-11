package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nicklaw5/helix/v2"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/data"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/syncx"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/utils"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

/*
Adding channel for autosr: (nightbot just link rewardsmap)
1) If using spotify, get credentials using cmd/spotifyoauth/main.go, and add them to tokens folder
*/

type SpotifyClientState struct {
	SpotifyClient *spotify.Client
	LastSkip      time.Time
	LastSongCmd   time.Time
	LastQueueCmd  time.Time
}

var spotifyStates = map[string]SpotifyClientState{}

var requesters = syncx.Map[string, string]{}

func PeriodicallyCleanRequesters(client *helix.Client, channel string) {
	for {
		time.Sleep(time.Hour)
		live, err := utils.ChannelIsLive(client, channel)
		if err != nil {
			continue
		}
		if !live {
			requesters = syncx.Map[string, string]{}
		}
	}
}

func StartupSpotify() {
	// setup app client
	// ctx := context.Background()
	// config := &clientcredentials.Config{
	// 	ClientID:     data.AppCfg.SpotifyID,
	// 	ClientSecret: data.AppCfg.SpotifySecret,
	// 	TokenURL:     spotifyauth.TokenURL,
	// }
	// token, err := config.Token(ctx)
	// if err != nil {
	// 	log.Fatalf("couldn't get token: %v", err)
	// }

	// httpClient := spotifyauth.New().Client(ctx, token)
	// appClient = spotify.New(httpClient)

	// Read auth and get auth
	files, err := os.ReadDir("tokens")
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		c := file.Name()
		if !strings.HasSuffix(c, ".json") {
			continue
		}
		jsonFile, err := os.Open("tokens/" + c)
		if err != nil {
			log.Println("No auth for", c)
			continue
		}
		defer jsonFile.Close()
		b, err := io.ReadAll(jsonFile)
		if err != nil {
			log.Println("Cannot read auth json", c)
			continue
		}
		var auth oauth2.Token
		err = json.Unmarshal(b, &auth)
		if err != nil {
			log.Println("json format error for auth", c)
			continue
		}

		// auth with spotify client
		userClient := spotify.New(spotifyauth.New(
			spotifyauth.WithRedirectURL("http://localhost:51337"),
			spotifyauth.WithScopes(
				spotifyauth.ScopeUserModifyPlaybackState,
				spotifyauth.ScopeUserReadCurrentlyPlaying,
				spotifyauth.ScopeUserReadPlaybackState,
				spotifyauth.ScopeUserReadRecentlyPlayed,
				spotifyauth.ScopeUserReadEmail,
			),
			spotifyauth.WithClientID(data.AppCfg.SpotifyID),
			spotifyauth.WithClientSecret(data.AppCfg.SpotifySecret),
		).Client(context.Background(), &auth))

		spotifyStates[strings.TrimSuffix(c, ".json")] = SpotifyClientState{
			SpotifyClient: userClient,
			LastSkip:      time.Now().Add(-10 * time.Second),
			LastSongCmd:   time.Now().Add(-10 * time.Second),
			LastQueueCmd:  time.Now().Add(-1 * time.Minute),
		}
		log.Println("Spotify client set for " + c)
	}
}

func SkipSongSpotify(irc *IRCConn, channel string, user string, permissionLevel int, brokenMessage []string) {
	var ok bool
	if _, ok = spotifyStates[channel]; !ok {
		irc.MsgChan <- Chat("Uh... "+data.AppCfg.TwitchAccount+" broken ...", channel, []string{})
		log.Println("Error: spotify state dne", channel)
		return
	}
	if spotifyStates[channel].SpotifyClient == nil {
		irc.MsgChan <- Chat("Uh... "+data.AppCfg.TwitchAccount+" broken ...", channel, []string{})
		log.Println("Error: Calling nil state.SpotifyClient fopr channel", channel)
		return
	}
	now := time.Now()
	if spotifyStates[channel].LastSkip.Add(time.Second * 10).After(now) {
		irc.MsgChan <- Chat("Don't skip song too quickly!", channel, []string{})
		return
	}
	live, err := utils.ChannelIsLive(irc.helixMainClient, strings.Trim(channel, "#"))
	if err != nil {
		irc.MsgChan <- Chat("Sorry I couldn't check if the broadcaster is live", channel, []string{})
		return
	}
	if !live {
		irc.MsgChan <- Chat("Broadcaster is not live you sillyy", channel, []string{})
		return
	}
	newState := spotifyStates[channel]
	newState.LastSkip = now
	spotifyStates[channel] = newState

	err = spotifyStates[channel].SpotifyClient.Next(context.Background())
	if err != nil {
		irc.MsgChan <- Chat("Failed to skip song "+err.Error(), channel, []string{})
		return
	}
	irc.MsgChan <- Chat("Skipped song", channel, []string{})
}

func CheckCurrentSongSpotify(irc *IRCConn, channel string, permissionLevel int, brokenMsg []string) {
	var ok bool
	var state SpotifyClientState
	if state, ok = spotifyStates[channel]; !ok {
		return
	}
	if state.SpotifyClient == nil {
		return
	}
	now := time.Now()
	if spotifyStates[channel].LastSongCmd.Add(time.Second * 10).After(now) {
		return
	}
	live, err := utils.ChannelIsLive(irc.helixMainClient, strings.Trim(channel, "#"))
	if err != nil {
		irc.MsgChan <- Chat("I couldn't check if the broadcaster is live", channel, []string{})
		return
	}
	if !live {
		irc.MsgChan <- Chat("Broadcaster is not live you sillyy", channel, []string{})
		return
	}
	queue, err := state.SpotifyClient.GetQueue(context.Background())
	if err != nil {
		irc.MsgChan <- Chat("Error: Couldn't check currently playing song "+err.Error(), channel, []string{})
		return
	}
	state.LastSongCmd = now
	from, _ := requesters.Load(queue.CurrentlyPlaying.ID.String())

	irc.MsgChan <- Chat(queue.CurrentlyPlaying.Name+" by "+queue.CurrentlyPlaying.Artists[0].Name+" Reqested by "+from, channel, []string{})

}

func ShowQueue(irc *IRCConn, channel string, user string, permissionLevel int, brokenMsg []string) {
	var ok bool
	var state SpotifyClientState
	if state, ok = spotifyStates[channel]; !ok {
		return
	}
	if state.SpotifyClient == nil {
		return
	}
	now := time.Now()
	if spotifyStates[channel].LastQueueCmd.Add(time.Minute).After(now) {
		return
	}
	live, err := utils.ChannelIsLive(irc.helixMainClient, strings.Trim(channel, "#"))
	if err != nil {
		irc.MsgChan <- Chat("I couldn't check if the broadcaster is live", channel, []string{})
		return
	}
	if !live {
		irc.MsgChan <- Chat("Broadcaster is not live you sillyy", channel, []string{})
		return
	}
	queue, err := state.SpotifyClient.GetQueue(context.Background())
	if err != nil {
		irc.MsgChan <- Chat("Error: Couldn't check currently playing song "+err.Error(), channel, []string{})
		return
	}
	if len(queue.Items) < 5 {
		irc.MsgChan <- Chat("Error: There is less than 5 songs in queue???", channel, []string{})
		return
	}
	state.LastQueueCmd = now
	msg := "Now: " + queue.CurrentlyPlaying.Name + " by " + queue.CurrentlyPlaying.Artists[0].Name + ", "
	for i, v := range queue.Items {
		if i > 4 {
			break
		}
		msg += "#" + strconv.Itoa(i+1) + ": " + v.Name + " by " + v.Artists[0].Name + ", "
	}
	irc.MsgChan <- Chat(strings.TrimSuffix(msg, ", "), channel, []string{})
}

func ProcessSongRequestSpotify(irc *IRCConn, channel string, user string, permissionLevel int, brokenMsg []string) {
	var ok bool
	var state SpotifyClientState
	if state, ok = spotifyStates[channel]; !ok {
		return
	}
	if state.SpotifyClient == nil {
		// log.Println("Error: Calling nil state.SpotifyClient fopr channel", channel)
		return
	}
	live, err := utils.ChannelIsLive(irc.helixMainClient, strings.Trim(channel, "#"))
	if err != nil {
		irc.MsgChan <- Chat("I couldn't check if the broadcaster is live", channel, []string{})
		return
	}
	if !live {
		irc.MsgChan <- Chat("Broadcaster is not live you sillyy", channel, []string{})
		return
	}
	TrackID := ""
	for _, s := range brokenMsg {
		if strings.Contains(s, "youtube.com") || strings.Contains(s, "youtu.be") {
			// Ignore youtube song requests
			irc.MsgChan <- Chat("Sorry I only support spotify!", channel, []string{})
			return
		}
		if strings.HasPrefix(s, "spotify:track:") || strings.HasPrefix(s, "https://open.spotify.com/track/") {
			if strings.HasPrefix(s, "https://open.spotify.com/track/") {
				TrackID = strings.TrimPrefix(s, "https://open.spotify.com/track/")
				if strings.Contains(TrackID, "?") {
					TrackID = TrackID[:strings.Index(TrackID, "?")]
				}
			} else {
				TrackID = strings.TrimPrefix(s, "spotify:track:")
			}
			break
		}
	}
	// text search
	if TrackID == "" {
		ctx := context.Background()
		result, err := state.SpotifyClient.Search(ctx, strings.Join(brokenMsg[1:], " "), spotify.SearchTypeTrack, spotify.Market(data.AppCfg.SpotifyRegion), spotify.Limit(1))
		if err != nil {
			irc.MsgChan <- Chat("Error when searching track "+err.Error(), channel, []string{})
			return
		}
		if len(result.Tracks.Tracks) > 0 {
			TrackID = result.Tracks.Tracks[0].ID.String()
		} else {
			irc.MsgChan <- Chat("No results found on Spotify", channel, []string{})
			return
		}
	}
	result, err := state.SpotifyClient.GetTrack(context.Background(), spotify.ID(TrackID))
	if err != nil {
		irc.MsgChan <- Chat("No results found on Spotify", channel, []string{})
		return
	}
	// Check for song length
	if result.Duration/1000 >= 360 { // 6 minutes
		irc.MsgChan <- Chat("Is this really a song if it's longer than 6 minutes???", channel, []string{})
		return
	}
	// Check if song is in queue already
	queue, err := state.SpotifyClient.GetQueue(context.Background())
	if err != nil {
		irc.MsgChan <- Chat("Error: Couldn't check if your song was already queued "+err.Error(), channel, []string{})
		return
	}
	if queue.CurrentlyPlaying.ID.String() == TrackID {
		irc.MsgChan <- Chat("It's the currently playing song you sillyy", channel, []string{})
		return
	}
	for _, v := range queue.Items {
		if v.ID.String() == TrackID {
			irc.MsgChan <- Chat("That song is already queued you sillyy", channel, []string{})
			return
		}
	}

	err = state.SpotifyClient.QueueSong(context.Background(), spotify.ID(TrackID))
	if err != nil {
		irc.MsgChan <- Chat("Error adding track to queue "+err.Error(), channel, []string{})
		return
	}
	user = utils.RawIRCUserToUsername(user)
	requesters.Store(TrackID, user)
	irc.MsgChan <- Chat(user+" added "+result.Name+" by "+result.Artists[0].Name+" to queue", channel, []string{})
}
