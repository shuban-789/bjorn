package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type TeamInfo struct {
	Number     int      `json:"number"`
	Name       string   `json:"name"`
	SchoolName string   `json:"schoolName"`
	Sponsors   []string `json:"sponsors"`
	Country    string   `json:"country"`
	State      string   `json:"state"`
	City       string   `json:"city"`
	RookieYear int      `json:"rookieYear"`
	Website    string   `json:"website"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
}

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
			teamAwards(ChannelID, teamNumber, session)
		default:
			session.ChannelMessageSend(ChannelID, "Unknown subcommand. Use 'stats' or 'awards'.")
		}
	} else {
		showTeamInfo(ChannelID, teamNumber, session)
	}
}

func fetchTeamInfo(teamNumber string) (*TeamInfo, error) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/teams/%s", teamNumber)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch info for Team %s: %v", teamNumber, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response for Team %s: %v", teamNumber, err)
	}

	var team TeamInfo
	err = json.Unmarshal(body, &team)
	if err != nil {
		return nil, fmt.Errorf("failed to parse info for Team %s: %v", teamNumber, err)
	}

	return &team, nil
}

// Default FTCScout API
func showTeamInfo(ChannelID string, teamNumber string, session *discordgo.Session) {
	team, err := fetchTeamInfo(teamNumber)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Info for Team %d (%s)", team.Number, team.Name),
		Description: fmt.Sprintf("**Team Number:** %d\n**School:** %s\n**City/State:** %s, %s\n**Rookie Year:** %d\n**Country:** %s",
			team.Number, team.SchoolName, team.City, team.State, team.RookieYear, team.Country),
		Color: 0x72cfdd,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Website",
				Value:  team.Website,
				Inline: true,
			},
			{
				Name:   "Sponsors",
				Value:  fmt.Sprintf("%v", team.Sponsors),
				Inline: true,
			},
		},
	}

	// HandleErr() can't be used for client-side response
	_, err = session.ChannelMessageSendEmbed(ChannelID, embed)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to send embed: %v", err))
	}
}

// Stats FTCScout API
func teamStats(ChannelID string, teamNumber string, session *discordgo.Session) {
	team, err := fetchTeamInfo(teamNumber)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

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

	body, err := io.ReadAll(resp.Body)
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
		Title:       fmt.Sprintf("Stats for Team %d (%s)", team.Number, team.Name),
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

	// HandleErr() can't be used for client-side response
	_, err = session.ChannelMessageSendEmbed(ChannelID, embed)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to send embed: %v", err))
	}
}

// Awards FTCScout API
func teamAwards(ChannelID string, teamNumber string, session *discordgo.Session) {
	team, err := fetchTeamInfo(teamNumber) // Reuse fetchTeamInfo to get the team name
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/teams/%s/awards", teamNumber)
	resp, err := http.Get(url)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to fetch awards for Team %s: %v", teamNumber, err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to read response for Team %s: %v", teamNumber, err))
		return
	}

	type TeamAward struct {
		Season       int    `json:"season"`
		EventCode    string `json:"eventCode"`
		TeamNumber   int    `json:"teamNumber"`
		Type         string `json:"type"`
		Placement    int    `json:"placement"`
		DivisionName string `json:"divisionName"`
		PersonName   string `json:"personName"`
		CreatedAt    string `json:"createdAt"`
		UpdatedAt    string `json:"updatedAt"`
	}

	var awards []TeamAward
	err = json.Unmarshal(body, &awards)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to parse awards for Team %s: %v", teamNumber, err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Awards for Team %d (%s)", team.Number, team.Name),
		Description: "Here are the awards this team has received:",
		Color:       0x72cfdd,
	}

	for _, award := range awards {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s (%d)", award.Type, award.Season),
			Value: fmt.Sprintf("Placement: %d\nEvent Code: %s", award.Placement, award.EventCode),
		})
	}

	if len(awards) == 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "No Awards",
			Value: "This team has not received any awards yet.",
		})
	}

	// HandleErr() can't be used for client-side response
	_, err = session.ChannelMessageSendEmbed(ChannelID, embed)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to send awards embed for Team %s: %v", teamNumber, err))
	}
}
