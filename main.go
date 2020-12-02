package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"netcitybot/netcity"
	"os"
	"strconv"
)

const (
	EnvKeyTgbotToken      = "BOT_API_TOKEN"
	EnvKeyTgbotChatId     = "BOT_CHAT_ID" // -1001402812566
	EnvKeyNetCityUsername = "NETCITY_USERNAME"
	EnvKeyNetCityPassword = "NETCITY_PASSWORD"
)

func main() {
	token := os.Getenv(EnvKeyTgbotToken)
	if token == "" {
		log.Fatal(fmt.Errorf("bot api token not found in env key: %s", EnvKeyTgbotToken))
	}

	chatId, err := strconv.ParseInt(os.Getenv(EnvKeyTgbotChatId), 10, 64)
	if err != nil {
		log.Fatal(fmt.Errorf("bot chat id error in env key %s: %s", EnvKeyTgbotChatId, err))
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	p := netcity.AuthParams{
		LoginType: 1,
		Cid:       2,
		Scid:      23,
		Pid:       -1,
		Cn:        3,
		Sft:       2,
		Sid:       66,
		Username:  os.Getenv(EnvKeyNetCityUsername),
		Password:  os.Getenv(EnvKeyNetCityPassword),
	}

	api := netcity.NewClientApi("https://netcity.eimc.ru", &p)
	assignments := map[int]netcity.DiaryAssignmentDetail{}
	go api.LoopPullingOrder(60, bot, chatId, &assignments, []int{76474, 76468})
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		var reply string
		switch update.Message.Command() {
		case "start":
			reply = "Привет. Я телеграм-бот по пересылке домашних заданий 6 \"Г\" класса из электронного дневника в чат"
		case "hello":
			reply = "И тебе привет."
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}
