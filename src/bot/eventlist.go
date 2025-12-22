// this functions similar to regions.go, but instead of being a hardcoded list,
// it fetches the data from the API and caches it in memory.
// it does this once every day.
package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shuban-789/bjorn/src/bot/util"
)

type EventData struct {
	Season int `json:"season"`
	Code  string `json:"code"`
	DivisionCode string `json:"divisionCode"`
	Name string `json:"name"`
	FieldCount int `json:"fieldCount"`
	Published bool `json:"published"`
	Type string `json:"type"`
	RegionCode string `json:"regionCode"`
	LeagueCode string `json:"leagueCode"`
	DistrictCode string `json:"districtCode"`
	Venue string `json:"venue"`
	Address string `json:"address"`
	Country string `json:"country"`
	State string `json:"state"`
	City string `json:"city"`
	Website string `json:"website"`
	LiveStreamUrl string `json:"liveStreamUrl"`
	Timezone string `json:"timezone"`
	StartDate string `json:"start"`
	EndDate string `json:"end"`
	ModifiedRules bool `json:"modifiedRules"`
	HasMatches bool `json:"hasMatches"`
}

type EventInfo struct {
	Code 	 string
	Name string
	Tokens []string // preprocess tokens for quick searching
	Region   RegionInfo
}

func (e EventInfo) GetSearchTokens() []string {
	return e.Tokens;
}

var cachedEventData map[string][]EventInfo = nil
var lastEventDataFetch time.Time

func FetchEvents() map[string][]EventInfo {
	if cachedEventData != nil && time.Since(lastEventDataFetch) < 24*time.Hour {
		return cachedEventData
	}
	
	fmt.Println(info("Fetching events data from API..."))
	api := "https://api.ftcscout.org/rest/v1/events/search/2025"

	resp, err := http.Get(api)
	if err != nil {
		fmt.Println(fail("Failed to fetch events data from API: %v", err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(fail("Events API returned status code: %d", resp.StatusCode))
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fail("Failed to read events response body: %v", err))
		return nil
	}

	var events []EventData
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Println(fail("Failed to parse events JSON response: %v", err))
		return nil
	}

	cachedEventData = make(map[string][]EventInfo)
	for _, event := range events {
		regionCode := event.RegionCode
		regionInfo := RegionInfo{
			Code: regionCode,
			Name: GetRegionFromCode(regionCode),
		}
		if regionInfo.Code == "" {
			continue
		}
		cachedEventData[regionInfo.Code] = append(cachedEventData[regionInfo.Code], EventInfo{
			Code: event.Code,
			Name: event.Name,
			Tokens: util.GenerateNormalizedTokens(event.Name),
			Region: regionInfo,
		})
	}
	lastEventDataFetch = time.Now()
	return cachedEventData
}

func startEventFetcher() {
	go func() {
		for {
			FetchEvents()
			fmt.Println(info("Event data refreshed"))
			time.Sleep(24 * time.Hour)
		}
	}();
}

func GetEventsData() map[string][]EventInfo{
	return cachedEventData
}

func SearchEventNames(query string, maxResults int, regionCode string) []EventInfo {
	eventsMap := GetEventsData()
	events := eventsMap[regionCode]
	return util.TokenizedSearch(events, query, maxResults)
}

func GetEventCodeFromName(name string, regionCode string) (code string, ok bool) {
	eventsMap := GetEventsData()
	events := eventsMap[regionCode]
	for _, event := range events {
		if event.Name == name {
			return event.Code, true
		}
	}
	return "", false
}

func init() {
	startEventFetcher()
}