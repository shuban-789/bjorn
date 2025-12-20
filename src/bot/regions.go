package bot

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
)

type RegionInfo struct {
	Code 	 string
	Name     RegionName
}

type RegionName struct {
	Name string
	Tokens []string // preprocess tokens for quick searching
}

// regions contains all the info in regionNames and regionCodes too, but we still keep
// those slices for easy access without having to loop through the structs each time.
var regions []RegionInfo = nil
var regionNames []RegionName = nil
var regionCodes []string = nil

func GetRegionsData() []RegionInfo {
	if regions != nil {
		return regions
	}
	file, err := os.Open("src/bot/data/regions.csv")
	if err != nil {
		fmt.Println(fail("Failed to open regions data file: %v", err))
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Println(fail("Failed to read regions data file via csv reader: %v", err))
	}

	for _, line := range lines {
		if len(line) != 2 {
			continue
		}
		regions = append(regions, RegionInfo{
			Code: line[0],
			Name: RegionName{
				Name: line[1],
				Tokens: strings.Fields(normalizeRegionName(line[1])),
			},
		})
	}
	return regions
}

func IsValidRegionCode(code string) bool {
	regions := GetRegionsData()
	for _, region := range regions {
		if region.Code == code {
			return true
		}
	}
	return false
}

func GetRegionName(code string) string {
	regions := GetRegionsData()
	for _, region := range regions {
		if region.Code == code {
			return region.Name.Name
		}
	}
	return ""
}

func GetAllRegionCodes() []string {
	if regionCodes != nil {
		return regionCodes
	}

	regions := GetRegionsData()
	codes := make([]string, 0, len(regions))

	for _, region := range regions {
		codes = append(codes, region.Code)
	}
	regionCodes = codes
	return regionCodes
}

func GetAllRegionNames() []RegionName {
	if regionNames != nil {
		return regionNames
	}

	regions := GetRegionsData()
	names := make([]RegionName, 0, len(regions))

	for _, region := range regions {
		names = append(names, region.Name)
	}
	regionNames = names
	return regionNames
}

func GetRegionCodeFromName(name string) string {
	regions := GetRegionsData()
	for _, region := range regions {
		if region.Name.Name == name {
			return region.Code
		}
	}
	return ""
}


// note: basically what I do here is just give a score to each region, then return the top n matches
// set maxResults to -1 to get all matches
func SearchRegionNames(query string, maxResults int) []string {
	names := GetAllRegionNames()
	if query == "" {
		results := make([]string, 0)
		for i, region := range names {
			if maxResults != -1 && i >= maxResults {
				break
			}
			results = append(results, region.Name)
		}
		return results
	}

	queryTokens := strings.Fields(normalizeRegionName(query))

	type ScoredRegion struct {
		region string
		score  int
	}
	scores := make([]ScoredRegion, 0)
	
	// 3 for loops is icky but it's like 90 regions so it's chill
	for _, region := range names {
		score := 0
		for _, queryToken := range queryTokens {
			for _, regionToken := range region.Tokens {
				// prioritize things starting w/ the query
				if strings.HasPrefix(regionToken, queryToken) {
					score += 2
				} else if strings.Contains(regionToken, queryToken) {
					score += 1
				}
			}
		}
		if score > 0 {
			scores = append(scores, ScoredRegion{
				region: region.Name,
				score:  score,
			})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	results := make([]string, 0)
	for i, scoredRegion := range scores {
		if maxResults != -1 && i >= maxResults {
			break
		}
		results = append(results, scoredRegion.region)
	}
	return results
}

// remove dashes + convert to lowercase for easier searching
func normalizeRegionName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", " "))
}

func init() {
	GetRegionsData()
	GetAllRegionCodes()
	GetAllRegionNames()
}
