package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/pagination"
	"github.com/shuban-789/bjorn/src/bot/search"
	"github.com/shuban-789/bjorn/src/bot/util"
)

var awardsPaginator *pagination.Paginator

var (
	awardsCache = util.NewCache(
		100,
		time.Hour*5,
		fetchTeamAwards,
	)
	awardsPerPage int = 5
)

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
	interactions.RegisterCommand(
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
							Name:        "team",
							Description: "The FTC team to look up.",
							Required:    true,
							Autocomplete: true,
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
							Name:        "team",
							Description: "The FTC team to look up.",
							Required:    true,
							Autocomplete: true,
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
							Name:        "team",
							Description: "The FTC team ID to look up.",
							Required:    true,
							Autocomplete: true,
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

					teamID := interactions.GetStringOption(sub.Options, "team")
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

	// ew go makes you put the period at the end or it assumes new line
	awardsPaginator = pagination.New("team;awards").
						AddExtraKey("teamNumber").
						OnCreate(generateAwardsEmbed).
						OnUpdate(updateAwardsEmbed).
						Register()

	searchAllTeamsAutocomplete := func(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice {
		results, err := search.SearchTeamNames(query, 25, "All")
		if err != nil {
			fmt.Println(util.Fail("Error searching team names: %v", err))
			return nil
		}

		choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(results))
		for _, team := range results {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%d %s", team.Number, team.Name),
				Value: fmt.Sprint(team.Number),
			})
		}
		return choices
	}

	interactions.RegisterAutocomplete("team/stats/team", searchAllTeamsAutocomplete)
	interactions.RegisterAutocomplete("team/awards/team", searchAllTeamsAutocomplete)
	interactions.RegisterAutocomplete("team/info/team", searchAllTeamsAutocomplete)
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

func fetchTeamAwards(teamNumber string) ([]TeamAward, error) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/teams/%s/awards", teamNumber)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch awards for Team %s: %v", teamNumber, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response for Team %s: %v", teamNumber, err)
	}

	var awards []TeamAward
	err = json.Unmarshal(body, &awards)
	if err != nil {
		return nil, fmt.Errorf("failed to parse awards for Team %s: %v", teamNumber, err)
	}
	return awards, nil
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

	teamNumberStr := fmt.Sprintf("%d", team.Number)
	
	awards, err := awardsCache.Get(teamNumberStr)
	
	initialState := pagination.PaginationState{
		TotalPages:  (len(awards) + awardsPerPage - 1) / awardsPerPage,
		CurrentPage: 0,
		ExtraData:   map[string]string{"teamNumber": fmt.Sprintf("%d", team.Number)},
	}
	err = awardsPaginator.Setup(session, i, channelID, initialState, team.Name)
	if err != nil {
		fmt.Println(util.Fail(err.Error()))
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to setup awards paginator for Team %s: %v", teamNumber, err))
		return
	}
}

func generateAwardsEmbed(state pagination.PaginationState, params ...any) (*discordgo.MessageEmbed, error) {
	teamNumberStr := state.ExtraData["teamNumber"]
	teamNum, err := strconv.Atoi(teamNumberStr)
	if err != nil {
		return &discordgo.MessageEmbed{}, err
	}
	
	teamName := params[0].(string)
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Awards for Team %d (%s)", teamNum, teamName),
		Description: "Here are the awards this team has received:",
		Color:       0x72cfdd,
	}

	return updateAwardsEmbed(state, embed)
}

// implements PageRenderer
func updateAwardsEmbed(state pagination.PaginationState, embed *discordgo.MessageEmbed) (*discordgo.MessageEmbed, error) {
	teamNumber := state.ExtraData["teamNumber"]
	
	awards, err := awardsCache.Get(teamNumber)
	if err != nil {
		return embed, err
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Page %d of %d", state.CurrentPage+1, state.TotalPages),
	}

	embed.Fields = []*discordgo.MessageEmbedField{}

	if len(awards) == 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "No Awards",
			Value: "This team has not received any awards yet.",
		})
		return embed, nil
	}

	for _, award := range awards[state.CurrentPage*awardsPerPage : min((state.CurrentPage+1)*awardsPerPage, len(awards))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s (%d)", award.Type, award.Season),
			Value: fmt.Sprintf("Placement: %d\nEvent Code: %s", award.Placement, award.EventCode),
		})
	}
	return embed, nil
}
