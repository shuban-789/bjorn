package bot

import (
	"github.com/bwmarrin/discordgo"
	"net/http"
	"strings"
	"fmt"
)

func teamcmd(ChannelID string, message *discordgo.MessageCreate, session *discordgo.Session) {
	args := strings.Fields(message.Content)
	if len(args) < 2 || args[0] != ">>team" {
		return
	}

	teamNumber := args[1]

	if len(args) > 2 {
		switch args[2] {
		case "stats":
			teamStats(ChannelID, teamNumber, session)
		case "awards":
			teamAwards(ChannelID, teamNumber, session)
		default:
			session.ChannelMessageSend(ChannelID, "Unknown subcommand. Available subcommands: stats, awards.")
		}
	} else {
		showTeamInfo(ChannelID, teamNumber, session)
	}
}

// Default FTCScout API
func showTeamInfo(ChannelID string, teamNumber string, session *discordgo.Session) {
	session.ChannelMessageSend(ChannelID, fmt.Sprintf("Team %s info:", teamNumber))
}

// Stats FTCScout API
func teamStats(ChannelID string, teamNumber string, session *discordgo.Session) {
	session.ChannelMessageSend(ChannelID, fmt.Sprintf("Stats for Team %s:", teamNumber))
}

// Awards FTCScout API
func teamAwards(ChannelID string, teamNumber string, session *discordgo.Session) {
	session.ChannelMessageSend(ChannelID, fmt.Sprintf("Awards for Team %s:", teamNumber))
}