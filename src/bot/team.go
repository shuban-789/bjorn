package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
)

// cache of team number to awards, used to reduce api calls
var awardsCache = map[int][]TeamAward{}
var awardsCacheMu = &sync.Mutex{} // btw this is bc the functions are goroutines so we don't want race conditions
var maxAwardsCacheSize int = 20   // I don't rlly want to use too much memory here
var awardsPerPage int = 5         // num awards per page, I think more than 10 is too much

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

func init() {
	RegisterCommand(
		&discordgo.ApplicationCommand{
			Name:        "team",
			Description: "Provides information about a specific FTC team.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "info",
					Description: "Show general team information.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "team_id",
							Description: "The FTC team ID to look up.",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stats",
					Description: "Show team statistics.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "team_id",
							Description: "The FTC team ID to look up.",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "awards",
					Description: "Show awards for a team.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "team_id",
							Description: "The FTC team ID to look up.",
							Required:    true,
						},
					},
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			var args []string
			if len(data.Options) > 0 {
				sub := data.Options[0]
				subName := sub.Name
				switch subName {
				case "info", "stats", "awards":
					if subName == "info" {
						subName = ""
					}

					teamID := getStringOption(sub.Options, "team_id")
					if teamID == "" {
						interactions.SendMessage(s, i, "", "Please provide a team number.")
						return
					}
					args = []string{teamID, subName}
				default:
					interactions.SendMessage(s, i, "", "Unknown subcommand for team.")
					return
				}
			}
			teamcmd(s, nil, i, args)
		},
	)

	RegisterComponentHandler("team;awards_p", func(s *discordgo.Session, ic *discordgo.InteractionCreate, data string) {
		parts := strings.Split(data, "_")
		teamNum, err1 := strconv.Atoi(parts[0])
		pageNum, err2 := strconv.Atoi(parts[1])
		totalPg, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			fmt.Print(errors.New(fail("Invalid pagination data! teamNum: %v, pageNum: %v, totalPg: %v", err1, err2, err3)))
			return
		}
		pageNum--
		if pageNum < 0 {
			pageNum = 0
		}
		embed := updateAwardsEmbed(teamNum, pageNum, ic.Message.Embeds[0])
		embeds := []*discordgo.MessageEmbed{embed}
		var id_prev string = fmt.Sprintf("team;awards_p %d_%d_%d", teamNum, pageNum, totalPg)
		var id_next string = fmt.Sprintf("team;awards_n %d_%d_%d", teamNum, pageNum, totalPg)
		components := []discordgo.MessageComponent{
			interactions.CreatePaginationButtons(totalPg, pageNum, id_prev, id_next),
		}
		s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		})
		// interactions.SendMessageComplex(s, ic, channelID, "", &components, &embeds)
	})

	RegisterComponentHandler("team;awards_n", func(s *discordgo.Session, ic *discordgo.InteractionCreate, data string) {
		parts := strings.Split(data, "_")
		teamNum, err1 := strconv.Atoi(parts[0])
		pageNum, err2 := strconv.Atoi(parts[1])
		totalPg, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			fmt.Print(errors.New(fail("Invalid pagination data! teamNum: %v, pageNum: %v, totalPg: %v", err1, err2, err3)))
			return
		}

		pageNum++
		if pageNum >= totalPg {
			pageNum = totalPg - 1
		}
		embed := updateAwardsEmbed(teamNum, pageNum, ic.Message.Embeds[0])
		embeds := []*discordgo.MessageEmbed{embed}
		var id_prev string = fmt.Sprintf("team;awards_p %d_%d_%d", teamNum, pageNum, totalPg)
		var id_next string = fmt.Sprintf("team;awards_n %d_%d_%d", teamNum, pageNum, totalPg)
		components := []discordgo.MessageComponent{
			interactions.CreatePaginationButtons(totalPg, pageNum, id_prev, id_next),
		}
		s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		})
		// interactions.SendMessageComplex(s, ic, channelID, "", &components, &embeds)
	})
}

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

// func teamcmd(channelID string, args []string, session *discordgo.Session, i *discordgo.InteractionCreate) {
func teamcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	channelID := interactions.GetChannelId(message, i)
	if len(args) < 1 {
		interactions.SendMessage(session, i, channelID, "Please provide a team number.")
		return
	}

	teamNumber := args[0]
	if len(args) > 1 && args[1] != "" {
		subCommand := args[1]
		switch subCommand {
		case "stats":
			teamStats(channelID, teamNumber, session, i)
		case "awards":
			teamAwards(channelID, teamNumber, session, i)
		default:
			interactions.SendMessage(session, i, channelID, "Unknown subcommand. Use 'stats' or 'awards'.")
		}
	} else {
		showTeamInfo(channelID, teamNumber, session, i)
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
func showTeamInfo(channelID string, teamNumber string, session *discordgo.Session, i *discordgo.InteractionCreate) {
	team, err := fetchTeamInfo(teamNumber)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Error: %v", err))
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

	interactions.SendEmbed(session, i, channelID, embed)
}

// Stats FTCScout API
func teamStats(channelID string, teamNumber string, session *discordgo.Session, i *discordgo.InteractionCreate) {
	team, err := fetchTeamInfo(teamNumber)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Error: %v", err))
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
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to fetch stats for Team %s: %v", teamNumber, err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to read response for Team %s: %v", teamNumber, err))
		return
	}

	var stats TeamStats
	err = json.Unmarshal(body, &stats)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to parse stats for Team %s: %v", teamNumber, err))
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
	interactions.SendEmbed(session, i, channelID, embed)
}

// Awards FTCScout API
func teamAwards(channelID string, teamNumber string, session *discordgo.Session, i *discordgo.InteractionCreate) {
	team, err := fetchTeamInfo(teamNumber) // Reuse fetchTeamInfo to get the team name
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Error: %v", err))
		return
	}

	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/teams/%s/awards", teamNumber)
	resp, err := http.Get(url)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to fetch awards for Team %s: %v", teamNumber, err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to read response for Team %s: %v", teamNumber, err))
		return
	}

	var awards []TeamAward
	err = json.Unmarshal(body, &awards)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to parse awards for Team %s: %v", teamNumber, err))
		return
	}

	totalPages := (len(awards) + awardsPerPage - 1) / awardsPerPage
	currentPage := 0

	saveAwardsToCache(team.Number, awards)
	embed := generateAwardsEmbed(team.Number, team.Name, currentPage)
	embeds := []*discordgo.MessageEmbed{embed}

	var id_prev string = fmt.Sprintf("team;awards_p %d_%d_%d", team.Number, currentPage, totalPages)
	var id_next string = fmt.Sprintf("team;awards_n %d_%d_%d", team.Number, currentPage, totalPages)

	components := []discordgo.MessageComponent{
		interactions.CreatePaginationButtons(totalPages, currentPage, id_prev, id_next),
	}

	interactions.SendMessageComplex(session, i, channelID, "", &components, &embeds)
}

func saveAwardsToCache(teamNumber int, awards []TeamAward) {
	awardsCacheMu.Lock()
	defer awardsCacheMu.Unlock()

	if len(awardsCache) >= maxAwardsCacheSize {
		// delete the first entry (I think this should be the oldest)
		for k := range awardsCache {
			delete(awardsCache, k)
			break
		}
	}
	awardsCache[teamNumber] = awards
}

func getAwardsFromCache(teamNumber int) ([]TeamAward, bool) {
	awardsCacheMu.Lock()
	defer awardsCacheMu.Unlock()
	awards, exists := awardsCache[teamNumber]
	return awards, exists
}

func generateAwardsEmbed(teamNumber int, teamName string, pageNumber int) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Awards for Team %d (%s)", teamNumber, teamName),
		Description: "Here are the awards this team has received:",
		Color:       0x72cfdd,
	}

	return updateAwardsEmbed(teamNumber, pageNumber, embed)
}

func updateAwardsEmbed(teamNumber, pageNumber int, embed *discordgo.MessageEmbed) *discordgo.MessageEmbed {
	awards, exists := getAwardsFromCache(teamNumber)
	if !exists {
		fmt.Println(errors.New(fail("Awards cache miss for team %d", teamNumber)))
		return embed
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Page %d of %d", pageNumber+1, (len(awards)+awardsPerPage-1)/awardsPerPage),
	}

	embed.Fields = []*discordgo.MessageEmbedField{}

	if len(awards) == 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "No Awards",
			Value: "This team has not received any awards yet.",
		})
		return embed
	}

	for _, award := range awards[pageNumber*awardsPerPage : min((pageNumber+1)*awardsPerPage, len(awards))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s (%d)", award.Type, award.Season),
			Value: fmt.Sprintf("Placement: %d\nEvent Code: %s", award.Placement, award.EventCode),
		})
	}
	return embed
}
