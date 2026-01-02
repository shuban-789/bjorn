package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/pagination"
	"github.com/shuban-789/bjorn/src/bot/util"
)

var (
	leadPaginator *pagination.Paginator[TeamRank]
	
	leadCache = util.NewCache(
		100,
		time.Hour*5,
		getLeaderboardInfo,
	)
)

func init() {
	interactions.RegisterCommand(
		&discordgo.ApplicationCommand{
			Name:        "lead",
			Description: "Display the leaderboard for a certain event.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "year",
					Description: "Year of the event (e.g., 2025).",
					Required:    true,
					Choices: interactions.FtcYearChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "region",
					Description: "The region the event is in (e.g., San Diego).",
					Required:    true,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "event",
					Description: "Event to look up.",
					Required:    true,
					Autocomplete: true,
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			year := interactions.GetStringOption(data.Options, "year")
			event := interactions.GetStringOption(data.Options, "event")
			if year == "" || event == "" {
				interactions.SendMessage(s, i, "", "Usage: /lead <year> <region> <event>")
				return
			}
			leadcmd(s, nil, i, []string{year, event})
		},
	)

	leadPaginator = pagination.New[TeamRank]("lead").
					ItemsPerPage(10).
					AddExtraKey("year").
					AddExtraKey("eventCode").
					OnUpdate(updateLeaderboard).
					WithDataGetter(func(state pagination.PaginationState) ([]TeamRank, error) {
						year := state.ExtraData["year"]
						eventCode := state.ExtraData["eventCode"]
						return leadCache.GetOrFetch(fmt.Sprintf("%s %s", year, eventCode))
					}).
					Register();
	
	// interactions.RegisterAutocomplete("lead/event", func(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice {
	// 	year := opts["year"]

	// })
	
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

func leadcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	channelId := interactions.GetChannelId(message, i)
	if len(args) < 2 {
		interactions.SendMessage(session, i, channelId, "Usage: >>lead <year> <eventCode>")
		return
	}

	err := leadPaginator.Setup(session, i, channelId, map[string]string{
		"year":      args[0],
		"eventCode": args[1],
	})
	if err != nil {
		interactions.SendMessage(session, i, channelId, fmt.Sprintf("Error sending leaderboard: %v", err))
	}
}

func getLeaderboardInfo(key string) ([]TeamRank, error) {
	splitted := strings.Fields(key)
	return fetchLeaderboard(splitted[0], splitted[1])
}

func fetchLeaderboard(year string, eventCode string) ([]TeamRank, error) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/teams", year, eventCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New(util.Fail("failed to fetch leaderboard: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(util.Fail("API returned status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(util.Fail("failed to read response: %v", err))
	}

	var leaderboard []map[string]interface{}
	if err := json.Unmarshal(body, &leaderboard); err != nil {
		return nil, errors.New(util.Fail("failed to parse JSON response: %v, body: %s", err, string(body)))
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

func updateLeaderboard(state pagination.PaginationState, data []TeamRank, previousEmbed *discordgo.MessageEmbed) (*discordgo.MessageEmbed, error) {
	year := state.ExtraData["year"]
	eventCode := state.ExtraData["eventCode"]
	return createLeaderboardEmbed(year, eventCode, data, state.CurrentPage+1, state.TotalPages), nil
}

func createLeaderboardEmbed(year string, eventCode string, teams []TeamRank, part int, totalParts int) *discordgo.MessageEmbed {
	title := fmt.Sprintf("%s %s Leaderboard", year, eventCode)
	if totalParts > 1 {
		title = fmt.Sprintf("%s (Part %d/%d)", title, part, totalParts)
	}

	embed := &discordgo.MessageEmbed{
		Title:  title,
		Color:  0x72cfdd,
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, team := range teams {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Rank %d", team.Rank),
			Value:  fmt.Sprintf("Team Number: %d", team.TeamNumber),
			Inline: false,
		})
	}

	return embed
}
