package main

import (
	"github.com/shuban-789/Gobot/src/bot"
	"os"
	"fmt"
)

func main() {
	BotToken := os.Getenv("DSC_BOT_TOKEN")
	if BotToken == "" {
		fmt.Println("\033[31m[FAIL]\033[0m No Bot Token found")
	}
	bot.Run(BotToken)
}