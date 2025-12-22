package bot

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/search"
	"github.com/shuban-789/bjorn/src/bot/util"
)

func init() {
	RegisterCommand(
		&discordgo.ApplicationCommand{
			Name:        "roleme",
			Description: "Assigns you a role based on your team ID.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "team",
					Description: "Your FTC team.",
					Required:    true,
					Autocomplete: true,
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			teamID := getStringOption(data.Options, "team")
			if teamID == "" {
				interactions.SendMessage(s, i, "", "Please provide a team number.")
				return
			}
			rolemeCmd(s, nil, i, []string{teamID})
		},
	)

	RegisterAutocompleteHandler("roleme", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()

		if len(data.Options) == 0 {
			return
		}
		fmt.Printf("roleme command autocomplete data: %+v\n", data)
		
		if !data.Options[0].Focused { return }

		teamQuery := data.Options[0].Value.(string)
		fmt.Printf("roleme autocomplete query: %s\n", teamQuery)
		results, err := search.SearchSDTeamNames(teamQuery, 25)
		if err != nil {
			fmt.Println(util.Fail("Error searching team names: %v", err))
			return
		}

		choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(results))
		for _, team := range results {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%s %s", team.TeamID, team.Name),
				Value: team.TeamID,
			})
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		})
		return
	})
}

func hash(ID string) string {
	hash := sha256.Sum256([]byte(ID))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

// func rolemeCmd(ChannelID string, args []string, session *discordgo.Session, guildId string, authorID string) {
func rolemeCmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	ChannelID := interactions.GetChannelId(message, i)
	guildId := interactions.GetGuildId(message, i)
	authorID := interactions.GetAuthorId(message, i)

	if len(args) != 1 {
		interactions.SendMessage(session, i, ChannelID, "Please provide a team number, and nothing more.")
		return
	}

	// shuban's blacklist code
	blacklistFile, err := os.Open("src/bot/data/blacklist.txt")
	if HandleErr(err) {
		interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't load the list of team names")
		return
	}
	defer blacklistFile.Close()

	blacklist := bufio.NewScanner(blacklistFile)
	for blacklist.Scan() {
		ban := blacklist.Text()
		hashedID := hash(authorID)
		if strings.Compare(ban, hashedID) == 0 {
			session.ChannelMessageSend(ChannelID, "Sorry, but you are banned from using this command.")
			return
		}
	}

	if err := blacklist.Err(); err != nil {
		HandleErr(err)
		interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't read the list of team names")
		return
	}

	teamNumber := args[0]
	teamName, err := search.GetSDTeamNameFromNumber(teamNumber)
	if err != nil {
		if err.Error() == "team number not found" {
			interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't find a team in San Diego with that ID competing in the DECODE:registered: season.")
		} else {
			interactions.SendMessage(session, i, ChannelID, "Sorry, an error occurred while searching for your team: "+err.Error())
		}
		return
	}
	
	// plan:
	// search for a role with the name "<team_id> <team_name>"
	// if that role exists, add it to the user
	// otherwise, create the role and then add it to the user

	var roleName string

	// get the roles
	roles, err := session.GuildRoles(guildId)
	if HandleErr(err) {
		interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't retrieve the roles in this server.")
		return
	}

	var roleID string
	for _, role := range roles {
		if strings.HasPrefix(role.Name, teamNumber) {
			roleID = role.ID
			roleName = role.Name
			break
		}
	}

	// if role doesn't exist
	if roleID == "" {
		roleName = teamNumber + " " + teamName
		color := 0x1ABC9C
		hoist := false
		mentionable := true

		roleInfo := &discordgo.RoleParams{
			Name:        roleName,
			Color:       &color,       // random color idk 0xfoodie lol
			Hoist:       &hoist,       // shown separately in member list
			Mentionable: &mentionable, // ya'll can ping teams
		}

		newRole, err := session.GuildRoleCreate(guildId, roleInfo)
		if HandleErr(err) {
			interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't create a new role.")
			return
		}

		roleID = newRole.ID
		interactions.SendMessage(session, i, ChannelID, "Creating a new role with name `"+roleName+"`.")

		interactions.SendMessage(session, i, ChannelID, "Do you want to set a color for the role? If yes, please provide a hex code. If not, type `no`.")

		session.AddHandlerOnce(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			if m.Author.ID != authorID || m.ChannelID != ChannelID {
				return
			}

			if strings.ToLower(m.Content) == "no" {
				interactions.SendMessage(session, i, ChannelID, "No color set for the role.")
				return
			}

			color, err := strconv.ParseInt(strings.TrimPrefix(m.Content, "#"), 16, 32)
			if err != nil {
				interactions.SendMessage(session, i, ChannelID, "Invalid hex code. No color set for the role.")
				return
			}

			colorInt := int(color)
			roleInfo.Color = &colorInt
			_, err = session.GuildRoleEdit(guildId, roleID, roleInfo)
			if HandleErr(err) {
				interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't set the color for the role.")
				return
			}

			interactions.SendMessage(session, i, ChannelID, "Color set for the role.")
		})
	}

	// add role to person
	err = session.GuildMemberRoleAdd(guildId, authorID, roleID)
	if HandleErr(err) {
		interactions.SendMessage(session, i, ChannelID, "Sorry, but I couldn't assign the role to you.")
		return
	}

	interactions.SendMessage(session, i, ChannelID, "You have been given the `"+roleName+"` role!")
}
