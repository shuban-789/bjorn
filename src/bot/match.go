package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Match struct {
	ID int `json:"id"`
}

type EventTracked struct {
	Year                 string
	EventCode            string
	UpdateChannelId      string
	LastUpdateTime       time.Time
	CachedMatches        []Match
	LastProcessedMatchId int
}

var eventsBeingTracked []EventTracked

func matchcmd(channelID string, args []string, session *discordgo.Session, guildId string, authorID string) {
	if len(args) < 1 {
		session.ChannelMessageSend(channelID, "Please provide a subcommand (e.g., 'info').")
		return
	}

	subCommand := args[0]

	switch subCommand {
	case "info":
		if len(args) < 4 {
			session.ChannelMessageSend(channelID, "Usage: `>>match info <year> <eventCode> <matchNumber>`")
			return
		}

		year := args[1]
		eventCode := args[2]
		matchNumber := args[3]

		getMatch(channelID, year, eventCode, matchNumber, session)
	case "eventstart":
		if len(args) < 3 {
			session.ChannelMessageSend(channelID, "Usage: `>>match eventstart <year> <eventCode>`")
			return
		}

		hasPerms, err := isAdmin(session, guildId, authorID)
		if err != nil {
			session.ChannelMessageSend(channelID, "Unable to check permissions of user.")
			return
		}

		if hasPerms {
			session.ChannelMessageSend(channelID, "You do not have permission to run this command.")
			return
		}

		year := args[1]
		eventCode := args[2]
		eventStart(channelID, year, eventCode, session)
	default:
		session.ChannelMessageSend(channelID, "Unknown subcommand. Available subcommands: `info`, `eventstart`")
	}
}

func eventStart(channelID string, year string, eventCode string, session *discordgo.Session) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to fetch match data: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		session.ChannelMessageSend(channelID, "That event does not exist!")
		return
	}

	eventsBeingTracked = append(eventsBeingTracked, EventTracked{
		Year:                 year,
		EventCode:            eventCode,
		UpdateChannelId:      channelID,
		LastUpdateTime:       time.Date(1, time.January, 1, 1, 1, 1, 1, time.Now().Location()), // hopefully will force an immediate update
		CachedMatches:        []Match{},
		LastProcessedMatchId: -100, // should probs force update
	})
	session.ChannelMessageSend(channelID, fmt.Sprintf("Started tracking matches for event %s in %s...", eventCode, year))
}

func getMatch(ChannelID string, year string, eventCode string, matchNumber string, session *discordgo.Session) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to fetch match data: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	var matches []struct {
		ID     int `json:"id"`
		Scores struct {
			Red struct {
				TotalPoints int `json:"totalPoints"`
			} `json:"red"`
			Blue struct {
				TotalPoints int `json:"totalPoints"`
			} `json:"blue"`
		} `json:"scores"`
		Teams []struct {
			Alliance   string `json:"alliance"`
			TeamNumber int    `json:"teamNumber"`
		} `json:"teams"`
	}
	err = json.Unmarshal(body, &matches)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to parse match data: %v", err))
		return
	}

	var selectedMatch struct {
		Scores struct {
			Red  int `json:"totalPoints"`
			Blue int `json:"totalPoints"`
		}
		RedTeams  []int
		BlueTeams []int
	}
	for _, match := range matches {
		if fmt.Sprintf("%d", match.ID) == matchNumber {
			selectedMatch.Scores.Red = match.Scores.Red.TotalPoints
			selectedMatch.Scores.Blue = match.Scores.Blue.TotalPoints
			for _, team := range match.Teams {
				if team.Alliance == "Red" {
					selectedMatch.RedTeams = append(selectedMatch.RedTeams, team.TeamNumber)
				} else if team.Alliance == "Blue" {
					selectedMatch.BlueTeams = append(selectedMatch.BlueTeams, team.TeamNumber)
				}
			}
			break
		}
	}

	winner := "Red Alliance"
	if selectedMatch.Scores.Blue > selectedMatch.Scores.Red {
		winner = "Blue Alliance"
	} else if selectedMatch.Scores.Blue == selectedMatch.Scores.Red {
		winner = "Draw"
	}

	color := 0xE02C44
	if strings.Compare(winner, "Blue Alliance") == 0 {
		color = 0x58ACEC
	} else if strings.Compare(winner, "Draw") == 0 {
		color = 0xE8E4EC
	}
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Match %s Results", eventCode, matchNumber),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Red Alliance Teams  🔴   \u200B",
				Value:  fmt.Sprintf("%v", selectedMatch.RedTeams),
				Inline: true,
			},
			{
				Name:   "Blue Alliance Teams  🔵   \u200B",
				Value:  fmt.Sprintf("%v", selectedMatch.BlueTeams),
				Inline: true,
			},
			{
				Name:   "\u200b",
				Value:  "",
				Inline: false,
			},
			{
				Name:   "Red Alliance Score  🔴   \u200B",
				Value:  fmt.Sprintf("** **%d", selectedMatch.Scores.Red),
				Inline: true,
			},
			{
				Name:   "Blue Alliance Score  🔵   \u200B",
				Value:  fmt.Sprintf("%d", selectedMatch.Scores.Blue),
				Inline: true,
			},
			{
				Name:   "\u200b",
				Value:  "",
				Inline: false,
			},
			{
				Name:   "Winner  🏆\u200B",
				Value:  winner,
				Inline: false,
			},
		},
		Color: color,
	}

	session.ChannelMessageSendEmbed(ChannelID, embed)
}

func eventUpdate(apiPollTime time.Duration, session *discordgo.Session) {
	for i := 0; i < len(eventsBeingTracked); i++ {
		var event EventTracked = eventsBeingTracked[i]

		var currentTime time.Time = time.Now()
		var notTimeElapsed bool = currentTime.Sub(event.LastUpdateTime) < apiPollTime
		eventsBeingTracked[i].LastUpdateTime = currentTime

		if notTimeElapsed { // time hasn't yet elapsed
			return
		}

		var url string = fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", event.Year, event.EventCode)
		resp, err := http.Get(url)
		if err != nil {
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to fetch match data: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			session.ChannelMessageSend(event.UpdateChannelId, "That event does not exist!")
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to read response: %v", err))
			return
		}

		err = json.Unmarshal(body, &event.CachedMatches)
		if err != nil {
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to parse match data: %v", err))
			return
		}

		var newMatches []int
		for _, match := range event.CachedMatches {
			if match.ID > event.LastProcessedMatchId {
				newMatches = append(newMatches, match.ID)
			}
		}

		if len(newMatches) > 0 {
			for _, matchID := range newMatches {
				getMatch(event.UpdateChannelId, event.Year, event.EventCode, fmt.Sprintf("%d", matchID), session)
			}

			event.LastProcessedMatchId = newMatches[len(newMatches)-1]
		} else {
			session.ChannelMessageSend(event.UpdateChannelId, "No new matches found in this interval.")
		}
	}
}
