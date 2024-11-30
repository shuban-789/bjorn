package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

func handleErr(err error) {
	if err != nil {
		fmt.Println("\033[31m[FAIL]\033[0m Error: %v", err)
	}
}

func Run(token string) {
	session, err := discordgo.New("Bot " + token)
	handleErr(err)
	session.AddHandler(tree)
	err = session.Open()
	defer session.Close()
	handleErr(err)
	fmt.Println("\033[32m[SUCCESS]\033[0m Bot is running")

	select {}
}

func tree(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}
	  
	switch {
		case strings.Contains(message.Content, "!help"):
			helpcmd(message.ChannelID, session)
		case strings.Contains(message.Content, "!ping"):
			pingcmd(message.ChannelID, session)
	}
}