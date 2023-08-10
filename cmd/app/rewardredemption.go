package main

import (
	"log"
	"os"

	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/api"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/data"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/utils"
)

/*
Add reward redemptions:
1) note down the uuid for the reward id
2) add it to rewardsMap and add the func to hook on it.
*/

var logreward = log.New(os.Stdout, "REWARD ", log.Ldate|log.Ltime)
var rewardsMap = map[string]func(*api.IRCConn, string, string, int, []string){
	"sr_spotify": api.ProcessSongRequestSpotify,
}

func handleRewards(irc *api.IRCConn, identity string, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	identityMap := utils.IdentityParser(identity)
	// logreward.Println(user+":", (*identityMap)["custom-reward-id"], brokenMessage)
	rewardID := (*identityMap)["custom-reward-id"]
	rewardName := ""

	// Add mechanism to detect which reward was claimed
	if data.AppCfg.TwitchSongReqestRewardID == rewardID {
		rewardName = "sr_spotify"
	}

	if f, ok := rewardsMap[rewardName]; ok {
		logreward.Println(user+":", rewardName, brokenMessage)
		f(irc, incomingChannel, user, permissionLevel, brokenMessage)
	}
}
