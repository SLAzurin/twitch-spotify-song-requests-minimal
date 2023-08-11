package main

import (
	"strings"
	"time"

	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/api"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/data"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/utils"
)

/*
Add commands under each channel here
*/

var AnyCommands = map[int]func(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string){
	1:    toggleAutoSR,
	2:    commandSkipSongSpotify,
	6:    commandProcessSongRequestSpotify,
	1001: commandSongSpotify,
	1002: commandCommands,
	1003: commandQueue,
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
	case "!song":
		fallthrough
	case "!currentsong":
		rCommandID = 1001
	case "!help":
		fallthrough
	case "!commands":
		rCommandID = 1002
	case "!q":
		fallthrough
	case "!queue":
		rCommandID = 1003
	}

	if f, ok := AnyCommands[rCommandID]; ok {
		f(irc, incomingChannel, user, permissionLevel, brokenMessage)
	}
}

var commandsLastUsed = time.Now().Add(-10 * time.Second)

func commandCommands(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	now := time.Now()
	if commandsLastUsed.Add(time.Second * 10).After(now) {
		return
	}
	irc.MsgChan <- api.Chat(utils.RawIRCUserToUsername(user)+" https://gist.github.com/SLAzurin/0288acb14791164b0e91844a515b049c", incomingChannel, []string{})
	commandsLastUsed = now
}

func commandSongSpotify(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	api.CheckCurrentSongSpotify(irc, incomingChannel, permissionLevel, brokenMessage)
}
func commandQueue(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	api.ShowQueue(irc, incomingChannel, user, permissionLevel, brokenMessage)
}

var toggleAutoSRCD = time.Now().Add(-10 * time.Second)

func toggleAutoSR(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	now := time.Now()
	if toggleAutoSRCD.Add(time.Second * 10).After(now) {
		return
	}
	toggleSR = !toggleSR
	toggleAutoSRCD = now

	if toggleSR {
		irc.MsgChan <- api.Chat("autosr is now on", incomingChannel, []string{})
	} else {
		irc.MsgChan <- api.Chat("autosr is now off", incomingChannel, []string{})
	}
}

func commandProcessSongRequestSpotify(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	if !toggleSR {
		return
	}
	api.ProcessSongRequestSpotify(irc, incomingChannel, user, permissionLevel, brokenMessage)
}

func commandSkipSongSpotify(irc *api.IRCConn, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	api.SkipSongSpotify(irc, incomingChannel, user, permissionLevel, brokenMessage)
}
