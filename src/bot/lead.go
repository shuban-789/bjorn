package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// Command structure:
// >>lead <season> <eventCode>

type TeamStats struct {
	Rank int `json:"rank"`
}

type TeamRank struct {
	Rank       int `json:"rank"`
	TeamNumber int `json:"teamNumber"`
}

func leadcmd(ChannelID string, args []string, session *discordgo.Session) {
	if len(args) < 1 {
		session.ChannelMessageSend(ChannelID, "Usage: >>lead <year> <eventCode>")
		return
	}

	year := args[0]
	eventCode := args[1]
	showLeaderboard(ChannelID, year, eventCode, session)
}

func fetchLeaderBoard(year string, eventCode string) ([]TeamRank, error) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/teams", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leaderboard: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("Raw API Response: %s\n", body)

	var leaderboard []map[string]interface{}
	if err := json.Unmarshal(body, &leaderboard); err != nil {
		return nil, fmt.Errorf("failed to parse leaderboard: %v", err)
	}

	fmt.Printf("Parsed Leaderboard Object: %#v\n", leaderboard)

	var ranks []TeamRank
	for _, team := range leaderboard {
		teamNumber, ok := team["teamNumber"].(int)
		if !ok {
			continue
		}

		stats, ok := team["stats"].(map[string]interface{})
		if !ok {
			continue
		}

		rank, ok := stats["rank"].(int32)
		if !ok {
			continue
		}

		fmt.Printf("Team Number: %d, Rank: %d\n", int(teamNumber), int(rank))

		ranks = append(ranks, TeamRank{
			Rank:       int(rank),
			TeamNumber: int(teamNumber),
		})
	}

	return ranks, nil
}

func showLeaderboard(ChannelID string, year string, eventCode string, session *discordgo.Session) {
	leaderboard, err := fetchLeaderBoard(year, eventCode)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Leaderboard for %s (%s)", year, eventCode),
		Description: "Here are the top teams and their ranks:",
		Color:       0x72cfdd,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for i, team := range leaderboard {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Rank %d", team.Rank),
			Value:  fmt.Sprintf("Team Number: %d", team.TeamNumber),
			Inline: false,
		})

		if i >= 24 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "More Teams",
				Value:  "Only the top 25 teams are shown.",
				Inline: false,
			})
			break
		}
	}

	_, err = session.ChannelMessageSendEmbed(ChannelID, embed)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to send embed: %v", err))
	}
}
