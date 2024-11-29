package main

import (
	"src/bot"
	"os"
)

func main() {
	BotToken := os.Getenv("DSC_BOT_TOKEN")
	if BotToken == "" {
		panic("No bot token provided")
	}
	bot.Run(BotToken)
}