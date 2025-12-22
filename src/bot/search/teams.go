package search

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/shuban-789/bjorn/src/bot/util"
)

type TeamInfo struct {
	Name   string
	TeamID string
	Tokens []string
}

func (t TeamInfo) GetSearchTokens() []string {
	return t.Tokens
}

var sdTeamNames []TeamInfo = make([]TeamInfo, 0, 0)

func loadSDTeamNames() ([]TeamInfo, error) {
	if len(sdTeamNames) > 0 {
		return sdTeamNames, nil
	}

	file, err := os.Open("src/bot/data/2025-26.txt")
	if err != nil {
		return nil, errors.New("couldn't open team names file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted := strings.Split(line, "*!*")
		if len(splitted) != 2 {
			continue
		}
		teamID := splitted[0]
		teamName := splitted[1]

		sdTeamNames = append(sdTeamNames, TeamInfo{
			Name:   teamName,
			TeamID: teamID,
			Tokens: util.GenerateNormalizedTokens(teamID + " " + teamName), // ["22105", "Runtime", "Terror"]
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("couldn't read team names file")
	}

	return sdTeamNames, nil
}

func SearchSDTeamNames(query string, maxResults int) ([]TeamInfo, error) {
	teams, err := loadSDTeamNames()
	if err != nil {
		return nil, err
	}
	return util.TokenizedSearch(teams, query, maxResults), nil
}

func GetSDTeamNameFromNumber(teamNumber string) (string, error) {
	teams, err := loadSDTeamNames()
	if err != nil {
		return "", err
	}

	for _, team := range teams {
		if team.TeamID == teamNumber {
			return team.Name, nil
		}
	}
	return "", errors.New("team number not found")
}
