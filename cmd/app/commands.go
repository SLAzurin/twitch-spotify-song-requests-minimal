package main

import (
	"strings"

	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/api"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/data"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/utils"
)

/*
Add commands under each channel here
*/

var AnyCommands = map[int]func(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string){
	1: toggleAutoSR,
	2: commandSkipSongSpotify,
	6: commandProcessSongRequestSpotify,
}

var toggleSR = true

func handleCommand(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	for _, v := range strings.Split(data.AppCfg.TwitchPermsOverrideUsers, ",") {
		if strings.EqualFold(utils.RawIRCUserToUsername(user), v) {
			permissionLevel = 6
		}
	}

	rCommandID := 0
	switch brokenMessage[0] {
	case "!sr":
		if permissionLevel < 1 { // sub
			// not enough perms
			return
		}
		rCommandID = 6
	case "!skip":
		fallthrough
	case "!next":
		if permissionLevel < 4 { // mod
			// not enough perms
			return
		}
		rCommandID = 2
	case "!autosr":
		fallthrough
	case "!togglesr":
		if permissionLevel < 4 { // mod
			// not enough perms
			return
		}
		rCommandID = 1
	}

	if f, ok := AnyCommands[rCommandID]; ok {
		f(irc, incomingChannel, user, permissionLevel, brokenMessage)
	}
}

func toggleAutoSR(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	toggleSR = !toggleSR

	if toggleSR {
		irc.MsgChan <- api.Chat("autosr is now on", incomingChannel, []string{})
	} else {
		irc.MsgChan <- api.Chat("autosr is now off", incomingChannel, []string{})
	}
}

func commandProcessSongRequestSpotify(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	api.ProcessSongRequestSpotify(irc, incomingChannel, permissionLevel, brokenMessage)
}

func commandSkipSongSpotify(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	api.SkipSongSpotify(irc, incomingChannel, user, permissionLevel, brokenMessage)
}
