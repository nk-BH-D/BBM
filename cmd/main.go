package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	BBM "github.com/nk-BH-D/BBM/FaI"
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panicf("ошибка создания бота: %v", err)
	}

	log.Printf("Авторизовались как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			BBM.HandleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			BBM.HandleCallback(bot, update.CallbackQuery)
		}
	}
}
