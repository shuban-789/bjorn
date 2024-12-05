package main

import (
	"fmt"
	"os"

	"github.com/shuban-789/bjorn/src/bot"
)

func main() {
	BotToken := os.Getenv("DSC_BOT_TOKEN")
	if BotToken == "" {
		fmt.Println("\033[31m[FAIL]\033[0m No Bot Token found")
		return
	}
	bot.Deploy(BotToken)
}
