package main

import (
	"runtime/debug"
	"strings"
	"time"

	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/api"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/data"
	"github.com/slazurin/twitch-spotify-song-requests-minimal/pkg/utils"
)

func bot(irc *api.IRCConn) {
	var err error
	part := ""
	for {
		var msg = make([]byte, 1024)
		var n int
		if n, err = irc.Conn.Read(msg); err != nil {
			irc.Log.Println("error when reading websocket msg", err)
			utils.LogToFile("error when reading websocket msg", err, string(debug.Stack()))
			if irc.Conn != nil {
				irc.Conn.Close()
				time.Sleep(2 * time.Second)
			}
			irc.ExitCh <- struct{}{}
			return
		}
		stringmsg := string(msg[:n])
		part += stringmsg
		index := strings.Index(part, "\r\n")
		for index != -1 {
			fullIRCMsg := part[:index]
			if fullIRCMsg != "" {
				go processIRC(irc, fullIRCMsg)
			}
			part = part[index+2:]
			index = strings.Index(part, "\r\n")
		}
	}

}

func processIRC(irc *api.IRCConn, incoming string) {
	breakdown := strings.Split(incoming, " ")
	identity := breakdown[0]
	user := ""
	if len(breakdown) > 1 {
		user = breakdown[1]
	}
	// incomingType := ""
	// if len(breakdown) > 1 {
	// 	incomingType = breakdown[2]
	// }
	incomingChannel := ""
	if len(breakdown) > 3 {
		incomingChannel = breakdown[3]
	}
	var brokenMessage []string
	if len(breakdown) > 4 {
		brokenMessage = breakdown[4:]
		brokenMessage[0] = brokenMessage[0][1:] // Removes colon from the first character of the full message
		if brokenMessage[0] == "ACTION" {
			brokenMessage = brokenMessage[1:]
		}
	}

	irc.Log.Println(incoming)

	switch {
	case incoming == ":tmi.twitch.tv RECONNECT":
		irc.ExitCh <- struct{}{}
		return
	case strings.HasPrefix(incoming, "PING"):
		irc.MsgChan <- strings.Replace(incoming, "PING", "PONG", 1)
		irc.Log.Println(strings.Replace(incoming, "PING", "PONG", 1))
	case strings.Contains(identity, "custom-reward-id="):
		handleRewards(irc, identity, incomingChannel, user, utils.GetPermissionLevel(utils.IdentityParser(identity), data.AppCfg.TwitchAccount), brokenMessage)
	case len(brokenMessage) > 0 && strings.HasPrefix(brokenMessage[0], "!"):
		handleCommand(irc, incomingChannel, user, utils.GetPermissionLevel(utils.IdentityParser(identity), data.AppCfg.TwitchAccount), brokenMessage)
	}
}
