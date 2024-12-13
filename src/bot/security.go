package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func isAdmin(s *discordgo.Session, guildID, userID string) (bool, error) {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get guild member: %w", err)
	}

	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			return false, fmt.Errorf("failed to get guild: %w", err)
		}
	}

	adminRoles := make(map[string]bool)
	for _, role := range guild.Roles {
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			adminRoles[role.ID] = true
		}
	}

	for _, roleID := range member.Roles {
		if adminRoles[roleID] {
			return true, nil
		}
	}

	return false, nil
}