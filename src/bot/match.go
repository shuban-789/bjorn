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
	ID              int    `json:"id"`
	hasBeenPlayed   bool   `json:"hasBeenPlayed"`
	actualStartTime string `json:"actualStartTime"`
}

func (m *Match) GetHasBeenPlayed() bool {
    return m.hasBeenPlayed
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
        event := &eventsBeingTracked[i]

        fmt.Printf("\033[33m[INFO]\033[0m Checking event: Year=%s, EventCode=%s, LastUpdateTime=%v\n",
            event.Year, event.EventCode, event.LastUpdateTime)

        if time.Since(event.LastUpdateTime) < apiPollTime {
            fmt.Printf("\033[33m[INFO]\033[0m Skipping update for event %s/%s (last updated %v)\n",
                event.Year, event.EventCode, event.LastUpdateTime)
            continue
        }

        event.LastUpdateTime = time.Now()

        url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", event.Year, event.EventCode)
        fmt.Printf("\033[33m[INFO]\033[0m Fetching matches from: %s\n", url)

        resp, err := http.Get(url)
        if err != nil {
            fmt.Printf("\033[31m[FAIL]\033[0m Failed to fetch match data for event %s/%s: %v\n",
                event.Year, event.EventCode, err)
            session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to fetch match data: %v", err))
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusNotFound {
            fmt.Printf("\033[33m[INFO]\033[0m Event %s/%s not found!\n", event.Year, event.EventCode)
            session.ChannelMessageSend(event.UpdateChannelId, "That event does not exist!")
            continue
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("\033[31m[FAIL]\033[0m Failed to read response for event %s/%s: %v\n",
                event.Year, event.EventCode, err)
            session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to read response: %v", err))
            continue
        }

        var matches []Match
        err = json.Unmarshal(body, &matches)
        if err != nil {
            fmt.Printf("\033[31m[ERROR]\033[0m Failed to parse match data for event %s/%s: %v\n",
                event.Year, event.EventCode, err)
            session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to parse match data: %v", err))
            continue
        }

        fmt.Printf("\033[33m[INFO]\033[0m Matches fetched for event %s/%s: %d matches found\n",
            event.Year, event.EventCode, len(matches))

        var newMatches []Match
        for _, match := range matches {
            fmt.Printf("\033[33m[INFO]\033[0m Checking match: ID=%d, hasBeenPlayed=%v\n",
                match.ID, match.GetHasBeenPlayed())

            if match.ID > event.LastProcessedMatchId && match.GetHasBeenPlayed() {
                newMatches = append(newMatches, match)
                fmt.Printf("\033[33m[INFO]\033[0m New match to process: ID=%d\n", match.ID)
            }
        }

        if len(newMatches) > 0 {
            for _, match := range newMatches {
                fmt.Printf("\033[33m[INFO]\033[0m Processing match: ID=%d\n", match.ID)
                getMatch(event.UpdateChannelId, event.Year, event.EventCode, fmt.Sprintf("%d", match.ID), session)
            }
            event.LastProcessedMatchId = newMatches[len(newMatches)-1].ID
            fmt.Printf("\033[33m[INFO]\033[0m Updated LastProcessedMatchId for event %s/%s: %d\n",
                event.Year, event.EventCode, event.LastProcessedMatchId)
        } else {
            fmt.Println("\033[33m[INFO]\033[0m No new played matches found in this interval.")
        }
    }
}


func startEventUpdater(session *discordgo.Session, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				eventUpdate(interval, session)
			}
		}
	}()
}
