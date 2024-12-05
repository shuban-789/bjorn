package bot

import (
	"bufio"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func rolemeCmd(ChannelID string, args []string, session *discordgo.Session, guildId string, authorID string) {
	if len(args) != 1 {
		session.ChannelMessageSend(ChannelID, "Please provide a team number, and nothing more.")
		return
	}

	var teamName string = ""

	teamNumber := args[0]
	file, err := os.Open("src/bot/util/2024-25.txt")
	if HandleErr(err) {
		session.ChannelMessageSend(ChannelID, "Sorry, but I couldn't load the list of team names")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted := strings.Split(line, "*!*")
		if splitted[0] == teamNumber {
			teamName = splitted[1]
		}
	}

	if err := scanner.Err(); err != nil {
		HandleErr(err)
		session.ChannelMessageSend(ChannelID, "Sorry, but I couldn't read the list of team names")
		return
	}

	if teamName == "" {
		session.ChannelMessageSend(ChannelID, "Sorry, but I couldn't find a team in San Diego with that ID competing in the INTO THE DEEP:registered: season.")
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
		session.ChannelMessageSend(ChannelID, "Sorry, but I couldn't retrieve the roles in this server.")
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
			session.ChannelMessageSend(ChannelID, "Sorry, but I couldn't create a new role.")
			return
		}

		roleID = newRole.ID
		session.ChannelMessageSend(ChannelID, "Creating a new role with name `"+roleName+"`.")
	}

	// add role to person
	err = session.GuildMemberRoleAdd(guildId, authorID, roleID)
	if HandleErr(err) {
		session.ChannelMessageSend(ChannelID, "Sorry, but I couldn't assign the role to you.")
		return
	}

	session.ChannelMessageSend(ChannelID, "You have been given the `"+roleName+"` role!")
}
