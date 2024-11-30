package bot

import ( 
	"github.com/bwmarrin/discordgo"
	"github.com/go-ping/ping"
	"time"
	"fmt"
)

func pingcmd(ChannelID string, session *discordgo.Session) {
	target := "google.com"
	pinger, err := ping.NewPinger(target)
	handleErr(err)

	pinger.Count = 1
	pinger.Timeout = 5 * time.Second

	err = pinger.Run()
	handleErr(err)

	stats := pinger.Statistics()
	session.ChannelMessageSend(ChannelID, fmt.Sprintf("🏓Pong! %vms", stats.AvgRtt.Milliseconds()))
}