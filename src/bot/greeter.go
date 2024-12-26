package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func memberJoinListener(session *discordgo.Session, event *discordgo.GuildMemberAdd) {
	channel, err := session.UserChannelCreate(event.User.ID)
	if err != nil {
		fmt.Println("\033[31m[FAIL]\033[0m Failed to create DM channel:", err)
		return
	}

	helpcmd(channel.ID, session)
}
