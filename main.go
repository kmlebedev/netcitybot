package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

const (
	EnvKeyTgbotToken      = "BOT_API_TOKEN"
	EnvKeyTgbotChatId     = "BOT_CHAT_ID"    // -1001402812566
	EnvKeyNetCitySchool   = "NETCITY_SCHOOL" // МБОУ СОШ №16
	EnvKeyNetCityUsername = "NETCITY_USERNAME"
	EnvKeyNetCityPassword = "NETCITY_PASSWORD"
	EnvKeyNetCityUrl      = "NETCITY_URL"         // https://netcity.eimc.ru"
	EnvKeyNetStudentIds   = "NETCITY_STUDENT_IDS" // 76474,76468
	EnvKeyYearId          = "NETCITY_YEAR_ID"
	EnvKeyRedisAddress    = "REDIS_ADDRESS"
	EnvKeyRedisDB         = "REDIS_DB"
	EnvKeyRedisPassword   = "REDIS_PASSWORD"
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
	yearId, err := strconv.Atoi(os.Getenv(EnvKeyYearId))
	if err != nil {
		log.Fatal(fmt.Errorf("netcity year id error in env key %s: %s", EnvKeyYearId, err))
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	redisOpt := redis.Options{
		Addr:     os.Getenv(EnvKeyRedisAddress),
		Password: os.Getenv(EnvKeyRedisPassword),
	}
	//
	var rdb *redis.Client
	if redisOpt.Password != "" {
		if db, err := strconv.Atoi(os.Getenv(EnvKeyRedisDB)); err != nil {
			redisOpt.DB = db
		}
		rdb = redis.NewClient(&redisOpt)
	}
	api := netcity.NewClientApi(&netcity.Config{
		Url:      os.Getenv(EnvKeyNetCityUrl),
		School:   os.Getenv(EnvKeyNetCitySchool),
		Username: os.Getenv(EnvKeyNetCityUsername),
		Password: os.Getenv(EnvKeyNetCityPassword),
	})
	assignments := map[int]netcity.DiaryAssignmentDetail{}
	var studentIds []int
	for _, strId := range strings.Split(strings.TrimSpace(os.Getenv(EnvKeyNetStudentIds)), ",") {
		if id, err := strconv.Atoi(strings.Trim(strId, " ")); err != nil {
			studentIds = append(studentIds, id)
		}
	}
	go api.LoopPullingOrder(60, bot, chatId, yearId, rdb, &assignments, &studentIds)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
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
		if _, err := bot.Send(msg); err != nil {
			log.Error(err)
		}
	}
}
