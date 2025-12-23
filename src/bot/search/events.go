// this functions similar to regions.go, but instead of being a hardcoded list,
// it fetches the data from the API and caches it in memory.
// it does this once every day.
package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
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
	Start string `json:"start"`
	End string `json:"end"`
	Ongoing bool `json:"ongoing"`
	ModifiedRules bool `json:"modifiedRules"`
	HasMatches bool `json:"hasMatches"`
}

type EventInfo struct {
	Code 	 string
	Name     string
	Tokens   []string // preprocess tokens for quick searching
	Region   RegionInfo
	End      string
	Timezone string
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
	
	fmt.Println(util.Info("Fetching events data from API..."))
	api := "https://api.ftcscout.org/rest/v1/events/search/2025"

	resp, err := http.Get(api)
	if err != nil {
		fmt.Println(util.Fail("Failed to fetch events data from API: %v", err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(util.Fail("Events API returned status code: %d", resp.StatusCode))
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(util.Fail("Failed to read events response body: %v", err))
		return nil
	}

	var events []EventData
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Println(util.Fail("Failed to parse events JSON response: %v", err))
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
			End: event.End,
			Timezone: event.Timezone,
		})
	}
	lastEventDataFetch = time.Now()
	return cachedEventData
}

func startEventFetcher() {
	go func() {
		for {
			FetchEvents()
			fmt.Println(util.Info("Event data refreshed"))
			time.Sleep(24 * time.Hour)
		}
	}();
}

func GetEventsData() map[string][]EventInfo{
	return cachedEventData
}

func SearchEventNames(query string, maxResults int, regionCode string, includeFinishedEvents bool) []EventInfo {
	eventsMap := GetEventsData()
	events := eventsMap[regionCode]

	searchResults := util.TokenizedSearch(events, query, maxResults)

	// sort so that the earlier events are first in the list
	sort.Slice(searchResults, func(i, j int) bool {
		return EventEndsBefore(searchResults[i].End, searchResults[j].End)
	})

	if includeFinishedEvents {
		return searchResults
	}

	filteredResults := make([]EventInfo, 0, len(searchResults))
	for _, event := range searchResults {
		ended, err := EventHasEnded(event)
		if err != nil {
			continue
		}
		if !ended {
			filteredResults = append(filteredResults, event)
		}
	}
	return filteredResults
}

func EventEndsBefore(endA string, endB string) bool {
	layout := "2006-01-02"
	timeA, errA := time.Parse(layout, endA)
	timeB, errB := time.Parse(layout, endB)
	if errA != nil || errB != nil {
		return false
	}
	return timeA.Before(timeB)
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

func FetchEventData(year, eventCode string) (EventData, error) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		return EventData{}, fmt.Errorf("failed to fetch match data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return EventData{}, fmt.Errorf("that event does not exist!")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EventData{}, fmt.Errorf("failed to read response: %v", err)
	}

	var eventData EventData
	err = json.Unmarshal(body, &eventData)
	if err != nil {
		return EventData{}, fmt.Errorf("failed to parse event details: %v", err)
	}

	return eventData, nil
}

func GetEventStartEndTime(eventData EventData, today time.Time, location *time.Location) (time.Time, time.Time, error) {
	layout := "2006-01-02"
	startTime, err := time.Parse(layout, eventData.Start)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse event start time: %v", err)
	}	

	endTime, err := time.Parse(layout, eventData.End)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse event end time: %v", err)
	}	

	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 8, 0, 0, 0, location)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 17, 0, 0, 0, location)

	// If we start after 8, set it to be scheduled in the future so it doesn't error out
	if startTime.Year() == today.Year() && startTime.YearDay() == today.YearDay() && today.Hour() >= 8 {
		startTime = today.Add(5 * time.Minute)
	}	

	return startTime, endTime, nil
}

func GetEventEndTime(eventInfo EventInfo, today time.Time, location *time.Location) (time.Time, error) {
	layout := "2006-01-02"

	endTime, err := time.Parse(layout, eventInfo.End)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse event end time: %v", err)
	}	

	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 17, 0, 0, 0, location)
	return endTime, nil
}

func EventHasEnded(eventInfo EventInfo) (bool, error) {
	location, err := time.LoadLocation(eventInfo.Timezone)
	if err != nil {
		return false, fmt.Errorf("Error loading event timezone: %v", err)
	}

	today := time.Now().In(location)
	endTime, err := GetEventEndTime(eventInfo, today, location)
	if err != nil {
		return false, err
	}

	return endTime.Before(today), nil
}
