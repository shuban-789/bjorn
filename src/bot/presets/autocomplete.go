package presets

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/search"
	"github.com/shuban-789/bjorn/src/bot/util"
)

func RegionAutocomplete(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice {
	resultRegions := search.SearchRegionNames(query, 25)
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(resultRegions))
	for _, region := range resultRegions {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  region.Name,
			Value: region.Code,
		})
	}
	return choices
}

// note: requires region to be a parameter for the command
func EventAutocomplete(includeFinishedEvents bool) interactions.AutocompleteProvider {
	return func(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice {
		regionCode := opts["region"]
		if regionCode == "" {
			return nil
		}
		results := search.SearchEventNames(query, 25, regionCode, false)
		choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(results))
		for _, event := range results {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  event.Name,
				Value: event.Code,
			})
		}
		return choices
	}
}

func TeamsAutocomplete(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice {
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