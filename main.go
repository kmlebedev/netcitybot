package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
	var api *netcity.ClientApi
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-done
		log.Info("exiting")
		os.Exit(0)
	}()
	go func() {
		sig := <-sigs
		if api != nil {
			api.Logout()
		}
		log.Info(sig)
		done <- true
	}()

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

	redisOpt := redis.Options{
		Addr:     os.Getenv(EnvKeyRedisAddress),
		Password: os.Getenv(EnvKeyRedisPassword),
	}

	var rdb *redis.Client
	if redisOpt.Password != "" {
		if db, err := strconv.Atoi(os.Getenv(EnvKeyRedisDB)); err != nil {
			redisOpt.DB = db
		}
		rdb = redis.NewClient(&redisOpt)
	}
	api = netcity.NewClientApi(&netcity.Config{
		Url:      os.Getenv(EnvKeyNetCityUrl),
		School:   os.Getenv(EnvKeyNetCitySchool),
		Username: os.Getenv(EnvKeyNetCityUsername),
		Password: os.Getenv(EnvKeyNetCityPassword),
	})
	assignments := map[int]netcity.DiaryAssignmentDetail{}
	var pullStudentIds []int
	for _, strId := range strings.Split(strings.TrimSpace(os.Getenv(EnvKeyNetStudentIds)), ",") {
		if id, err := strconv.Atoi(strings.Trim(strId, " ")); err == nil {
			pullStudentIds = append(pullStudentIds, id)
		}
	}
	currentyYearId, err := strconv.Atoi(os.Getenv(EnvKeyYearId))
	if err != nil || currentyYearId == 0 {
		if currentyYearId, err = api.GetCurrentyYearId(); err != nil || currentyYearId == 0 {
			api.Logout()
			log.Fatalf("netcity year id error: %+v", err)
		}
	}
	if err := api.GetClasses(); err != nil {
		api.Logout()
		log.Fatalf("netcity get classes: %+v", err)
	}
	if err := api.GetAllStudents(); err != nil || len(api.Students) == 0 {
		api.Logout()
		log.Fatalf("netcity get all students:%+v %+v", len(api.Students), err)
	}

	log.Infof("Sync years: %d, classes: %d, students: %d", len(api.Years), len(api.Classes), len(api.Students))
	go api.LoopPullingOrder(60, bot, chatId, currentyYearId, rdb, &assignments, &pullStudentIds)
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
