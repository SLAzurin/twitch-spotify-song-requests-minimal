package api

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nicklaw5/helix/v2"

	"golang.org/x/net/websocket"
)

/*
Main entrypoint
*/

type IRCConn struct {
	Conn            *websocket.Conn
	MsgChan         chan string
	ExitCh          chan struct{}
	Log             *log.Logger
	helixMainClient *helix.Client
}

var instance *IRCConn

var host = "wss://irc-ws.chat.twitch.tv"

func (irc *IRCConn) IRCMessage(s string) {
	websocket.Message.Send(irc.Conn, s)
}

func Chat(s string, channel string, whitelist []string) string {
	// whitelist is a remnant of other code from another project
	return "PRIVMSG " + channel + " :/me " + s
}

func RunIRC(app func(*IRCConn), credentials struct {
	Password        string
	Nickname        string
	Channel         string
	HelixMainClient *helix.Client
}) (*IRCConn, error) {
	connectRetries := 0
	var err error
	var rawConn *websocket.Conn
	instance = &IRCConn{
		ExitCh:          make(chan struct{}, 1),
		MsgChan:         make(chan string),
		Log:             log.New(os.Stdout, "IRC ", log.Ldate|log.Ltime),
		helixMainClient: credentials.HelixMainClient,
	}

	for instance.Conn == nil {
		time.Sleep(time.Second * time.Duration(connectRetries))
		rawConn, err = websocket.Dial(host, "", "http://localhost/")
		if err != nil {
			if connectRetries > 128 {
				instance.Log.Println("Last retry took 128s and still didn't reconnect")
				instance.Log.Println("Force closing")
				instance.ExitCh <- struct{}{}
				return instance, errors.New("failed to retry 128s later")
			}
			instance.Log.Println("Failed to connect", err)
			instance.Log.Println("Retrying")
			if connectRetries == 0 {
				connectRetries = 1
			} else {
				connectRetries *= 2
			}
			continue
		}
		connectRetries = 0
		instance.Conn = rawConn
	}

	go func() {
		for s := range instance.MsgChan {
			if !strings.HasPrefix(s, "PASS") {
				instance.Log.Println("Send: " + s)
			} else {
				instance.Log.Println("Send: STUFF THAT SHOULD NOT BE LOGGED")
			}
			instance.IRCMessage(s)
		}
	}()

	// Login IRC
	instance.MsgChan <- "CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands"
	instance.MsgChan <- "PASS oauth:" + credentials.Password
	instance.MsgChan <- "NICK " + credentials.Nickname
	instance.MsgChan <- "JOIN " + credentials.Channel

	go app(instance)

	return instance, nil
}
