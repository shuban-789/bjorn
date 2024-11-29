package bot

import ( 
	"bwmarrin/discordgo"
	"github.com/go-ping/ping"
)

func ping(ChannelID string) {
	target := "google.com"
	pinger, err := ping.NewPinger(target)
	handleErr(err)

	pinger.Count = 1
	pinger.Timeout = 5 * time.Second

	err = pinger.Run()
	handleErr(err)

	stats := pinger.Statistics()
	fmt.Printf("🏓Pong! %v\n", stats.AvgRtt.Milliseconds())
}