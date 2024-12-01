package bot

import (
	"github.com/bwmarrin/discordgo"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

func teamcmd(ChannelID string, args []string, session *discordgo.Session) {
    if len(args) < 1 {
        session.ChannelMessageSend(ChannelID, "Please provide a team number.")
        return
    }

    teamNumber := args[0]
    if len(args) > 1 {
        subCommand := args[1]
        switch subCommand {
        case "stats":
            teamStats(ChannelID, teamNumber, session)
        case "awards":
            session.ChannelMessageSend(ChannelID, fmt.Sprintf("Awards for Team %s are not yet implemented.", teamNumber))
        default:
            session.ChannelMessageSend(ChannelID, "Unknown subcommand. Use 'stats' or 'awards'.")
        }
    } else {
        session.ChannelMessageSend(ChannelID, "Please specify a subcommand (e.g., 'stats', 'awards').")
    }
}

// Default FTCScout API
func showTeamInfo(ChannelID string, teamNumber string, session *discordgo.Session) {
	session.ChannelMessageSend(ChannelID, fmt.Sprintf("Team %s info:", teamNumber))
}

// Stats FTCScout API
func teamStats(ChannelID string, teamNumber string, session *discordgo.Session) {
	type TeamStats struct {
		Season int `json:"season"`
		Number int `json:"number"`
		Tot    struct {
			Value float64 `json:"value"`
			Rank  int     `json:"rank"`
		} `json:"tot"`
		Auto struct {
			Value float64 `json:"value"`
			Rank  int     `json:"rank"`
		} `json:"auto"`
		Dc struct {
			Value float64 `json:"value"`
			Rank  int     `json:"rank"`
		} `json:"dc"`
		Eg struct {
			Value float64 `json:"value"`
			Rank  int     `json:"rank"`
		} `json:"eg"`
		Count int `json:"count"`
	}

	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/teams/%s/quick-stats", teamNumber)
	resp, err := http.Get(url)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to fetch stats for Team %s: %v", teamNumber, err))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to read response for Team %s: %v", teamNumber, err))
		return
	}

	var stats TeamStats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to parse stats for Team %s: %v", teamNumber, err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Stats for Team %d", stats.Number),
		Color:       0x72cfdd,
		Description: fmt.Sprintf("Season: %d", stats.Season),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Total",
				Value:  fmt.Sprintf("%.2f (Rank: %d)", stats.Tot.Value, stats.Tot.Rank),
				Inline: false,
			},
			{
				Name:   "Auto",
				Value:  fmt.Sprintf("%.2f (Rank: %d)", stats.Auto.Value, stats.Auto.Rank),
				Inline: false,
			},
			{
				Name:   "DC",
				Value:  fmt.Sprintf("%.2f (Rank: %d)", stats.Dc.Value, stats.Dc.Rank),
				Inline: false,
			},
			{
				Name:   "EG",
				Value:  fmt.Sprintf("%.2f (Rank: %d)", stats.Eg.Value, stats.Eg.Rank),
				Inline: false,
			},
			{
				Name:   "Count",
				Value:  fmt.Sprintf("%d", stats.Count),
				Inline: false,
			},
		},
	}

	session.ChannelMessageSendEmbed(ChannelID, embed)
}

// Awards FTCScout API
func teamAwards(ChannelID string, teamNumber string, session *discordgo.Session) {
	session.ChannelMessageSend(ChannelID, fmt.Sprintf("Awards for Team %s:", teamNumber))
}