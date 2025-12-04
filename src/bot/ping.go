package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-ping/ping"
)

func pingcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate) {
	target := "google.com"
	pinger, err := ping.NewPinger(target)
	HandleErr(err)

	pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = 5 * time.Second

	err = pinger.Run()
	HandleErr(err)

	stats := pinger.Statistics()
	channelID := getChannelId(message, i)
	sendMessage(session, i, channelID, fmt.Sprintf("🏓 Pong! %vms", stats.AvgRtt.Milliseconds()))
}
