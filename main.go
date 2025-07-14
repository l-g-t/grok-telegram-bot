package main

import (
	"log"

	"github.com/dennislapchenko/grok-telegram-bot/bot"
	"github.com/dennislapchenko/grok-telegram-bot/config"
	"github.com/dennislapchenko/grok-telegram-bot/history"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	hist := history.NewChatHistory()

	b, err := bot.NewBot(&cfg, hist)
	if err != nil {
		log.Fatal(err)
	}

	b.Run()
}
