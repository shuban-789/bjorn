package bot

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/shuban-789/bjorn/src/bot/util"
)

type RegionInfo struct {
	Code 	 string
	Name string
	Tokens []string // preprocess tokens for quick searching
}

func (r RegionInfo) GetName() string {
	return r.Name
}

func (r RegionInfo) GetSearchTokens() []string {
	return r.Tokens
}

// regions contains all the info in regionNames and regionCodes too, but we still keep
// those slices for easy access without having to loop through the structs each time.
var regions []RegionInfo = nil
var regionNames []string = nil
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
			Name: line[1],
			Tokens: util.GenerateNormalizedTokens(line[1]),
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
			return region.Name
		}
	}
	return ""
}

func GetRegionFromCode(code string) string {
	regions := GetRegionsData()
	for _, region := range regions {
		if region.Code == code {
			return region.Name
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

func GetAllRegionNames() []string {
	if regionNames != nil {
		return regionNames
	}

	regions := GetRegionsData()
	names := make([]string, 0, len(regions))
	for _, region := range regions {
		names = append(names, region.Name)
	}
	regionNames = names
	return regionNames
}

func GetRegionCodeFromName(name string) string {
	regions := GetRegionsData()
	for _, region := range regions {
		if region.Name == name {
			return region.Code
		}
	}
	return ""
}


// note: basically what I do here is just give a score to each region, then return the top n matches
// set maxResults to -1 to get all matches
func SearchRegionNames(query string, maxResults int) []RegionInfo {
	return util.TokenizedSearch(regions, query, maxResults)
}

func init() {
	GetRegionsData()
	GetAllRegionCodes()
	GetAllRegionNames()
}
