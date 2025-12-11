package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-ping/ping"
	"github.com/shuban-789/bjorn/src/bot/interactions"
)

func init() {
	RegisterCommand(
		&discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Checks the bot's responsiveness.",
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			pingcmd(s, nil, i)
		},
	)
}

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
	channelID := interactions.GetChannelId(message, i)
	interactions.SendMessage(session, i, channelID, fmt.Sprintf("🏓 Pong! %vms", stats.AvgRtt.Milliseconds()))
}
