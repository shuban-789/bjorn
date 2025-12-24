package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shuban-789/bjorn/src/bot/util"
)

type TeamInfo struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
	Tokens []string
}

func (t TeamInfo) GetSearchTokens() []string {
	return t.Tokens
}

// map from region code to list of teams in that region
var teamNames map[string][]TeamInfo

func SearchTeamNames(query string, maxResults int, regionCode string) ([]TeamInfo, error) {
	teams := FetchTeams()
	return util.TokenizedSearch(teams[regionCode], query, maxResults), nil
}

func GetSDTeamNameFromNumber(teamNumber string) (string, error) {
	teams := FetchTeams()

	for _, team := range teams["USCASD"] { // hardcode to uscasd for now
		num, err := strconv.Atoi(teamNumber)
		if err != nil {
			return "", errors.New("invalid team number")
		}

		if team.Number == num {
			return team.Name, nil
		}
	}
	return "", errors.New("team number not found")
}

var lastTeamDataFetch time.Time

func FetchTeams() map[string][]TeamInfo {
	if teamNames != nil && time.Since(lastTeamDataFetch) < 10*24*time.Hour {
		return teamNames
	}

	if teamNames == nil {
		teamNames = make(map[string][]TeamInfo)
	}

	fmt.Println(util.Info("Fetching teams data from API..."))
	api := "https://api.ftcscout.org/rest/v1/teams/search?region="

	for _, region := range GetRegionsData() {
		fullApi := api + region.Code
		resp, err := http.Get(fullApi)
		if err != nil {
			fmt.Println(util.Fail("Failed to fetch teams data from API for region %s: %v", region.Code, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println(util.Fail("Teams API returned status code: %d for region %s", resp.StatusCode, region.Code))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(util.Fail("Failed to read teams response body: %v", err))
			return nil
		}

		var teams []TeamInfo
		if err := json.Unmarshal(body, &teams); err != nil {
			fmt.Println(util.Fail("Failed to parse teams JSON response: %v", err))
			return nil
		}

		for i := range teams {
			teams[i].Tokens = util.GenerateNormalizedTokens(strconv.Itoa(teams[i].Number) + " " + teams[i].Name)
		}

		teamNames[region.Code] = teams
	}
	lastTeamDataFetch = time.Now()
	return teamNames
}

func startTeamFetcher() {
	go func() {
		for {
			FetchTeams()
			fmt.Println(util.Info("Team data refreshed"))
			time.Sleep(20 * 24 * time.Hour) // refresh every 10 days, this lowk shouldn't change more than once a year
		}
	}();
}

func GetTeams() map[string][]TeamInfo{
	return teamNames
}

func GetTeamCodeFromName(name string, regionCode string) (code int, ok bool) {
	teamsMap := GetTeams()
	teams := teamsMap[regionCode]
	for _, team := range teams {
		if team.Name == name {
			return team.Number, true
		}
	}
	return -1, false
}

func init() {
	startTeamFetcher()
}	
