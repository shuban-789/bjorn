package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/bwmarrin/discordgo"
)

type TeamStats struct {
	Rank int `json:"rank"`
}

type TeamRank struct {
	Rank       int `json:"rank"`
	TeamNumber int `json:"teamNumber"`
}

type TeamRankSlice []TeamRank

func (s TeamRankSlice) Len() int {
	return len(s) 
}

func (s TeamRankSlice) Swap(i, j int) { 
	s[i], s[j] = s[j], s[i] 
}

func (s TeamRankSlice) Less(i, j int) bool {
	return s[i].Rank < s[j].Rank
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func leadcmd(ChannelID string, args []string, session *discordgo.Session) {
	if len(args) < 2 {
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
		return nil, fmt.Errorf("\033[31m[FAIL]\033[0m Failed to fetch leaderboard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("\033[31m[FAIL]\033[0m API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("\033[31m[FAIL]\033[0m Failed to read response: %v", err)
	}

	var leaderboard []map[string]interface{}
	if err := json.Unmarshal(body, &leaderboard); err != nil {
		return nil, fmt.Errorf("\033[31m[FAIL]\033[0m Failed to parse JSON response: %v, body: %s", err, string(body))
	}

	var ranks []TeamRank
	for _, team := range leaderboard {
		teamNumber, ok := team["teamNumber"].(float64)
		if !ok {
			continue
		}

		stats, ok := team["stats"].(map[string]interface{})
		if !ok {
			continue
		}

		rank, ok := stats["rank"].(float64)
		if !ok {
			continue
		}

		ranks = append(ranks, TeamRank{
			Rank:       int(rank),
			TeamNumber: int(teamNumber),
		})
	}

	sort.Sort(TeamRankSlice(ranks))
	return ranks, nil
}

func createLeaderboardEmbed(year string, eventCode string, teams []TeamRank, start int, end int, part int, totalParts int) *discordgo.MessageEmbed {
	title := fmt.Sprintf("%s %s Leaderboard", year, eventCode)
	if totalParts > 1 {
		title = fmt.Sprintf("%s (Part %d/%d)", title, part, totalParts)
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Color:       0x72cfdd,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for i := start; i < end && i < len(teams); i++ {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Rank %d", teams[i].Rank),
			Value:  fmt.Sprintf("Team Number: %d", teams[i].TeamNumber),
			Inline: false,
		})
	}

	return embed
}

func showLeaderboard(ChannelID string, year string, eventCode string, session *discordgo.Session) {
	leaderboard, err := fetchLeaderBoard(year, eventCode)
	if err != nil {
		errMsg := fmt.Sprintf("Error fetching leaderboard: %v", err)
		session.ChannelMessageSend(ChannelID, errMsg)
		return
	}

	if len(leaderboard) == 0 {
		msg := "No teams found in the leaderboard"
		session.ChannelMessageSend(ChannelID, msg)
		return
	}

	const teamsPerEmbed = 25
	totalTeams := len(leaderboard)
	numEmbeds := (totalTeams + teamsPerEmbed - 1) / teamsPerEmbed

	for i := 0; i < numEmbeds; i++ {
		start := i * teamsPerEmbed
		end := min((i + 1) * teamsPerEmbed, totalTeams)

		embed := createLeaderboardEmbed(year, eventCode, leaderboard, start, end, i+1, numEmbeds)
		
		_, err = session.ChannelMessageSendEmbed(ChannelID, embed)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to send embed part %d/%d: %v", i+1, numEmbeds, err)
			session.ChannelMessageSend(ChannelID, errMsg)
			return
		}
	}
}
