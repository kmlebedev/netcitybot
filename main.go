package main

import (
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
	EnvKeyNetCitySchool   = "NETCITY_SCHOOL" // МБОУ СОШ №53
	EnvKeyNetCityUsername = "NETCITY_USERNAME"
	EnvKeyNetCityPassword = "NETCITY_PASSWORD"
	EnvKeyNetCityUrl      = "NETCITY_URL"         // https://netcity.eimc.ru"
	EnvKeyNetStudentIds   = "NETCITY_STUDENT_IDS" // 76424,75468
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
		log.Fatal("bot api token not found in env key: %s", EnvKeyTgbotToken)
	}

	chatId, err := strconv.ParseInt(os.Getenv(EnvKeyTgbotChatId), 10, 64)
	if err != nil {
		log.Warning("bot chat id error in env key %s: %s", EnvKeyTgbotChatId, err)
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	// bot.Debug = true

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
	// sycn assignments details with attachments to telegram
	var pullStudentIds []int
	for _, strId := range strings.Split(strings.TrimSpace(os.Getenv(EnvKeyNetStudentIds)), ",") {
		if id, err := strconv.Atoi(strings.Trim(strId, " ")); err == nil {
			pullStudentIds = append(pullStudentIds, id)
		}
	}
	if chatId > 0 && len(pullStudentIds) > 0 {
		assignments := map[int]netcity.DiaryAssignmentDetail{}
		go api.LoopPullingOrder(60, bot, chatId, currentyYearId, rdb, &assignments, &pullStudentIds)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			if update.CallbackQuery == nil {
				continue
			}
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Error(err)
			}
			// And finally, send a message containing the data received.
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			if _, err := bot.Send(msg); err != nil {
				log.Error(err)
			}
		}
		if update.Message.Chat.ID != chatId {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		switch update.Message.Text {
		case "start":
			msg.Text = "Привет. Я телеграм-бот по пересылке домашних заданий 6 \"Г\" класса из электронного дневника в чат"
		case "hello":
			msg.Text = "И тебе привет."
		case "diary":
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("1"),
					tgbotapi.NewKeyboardButton("2"),
					tgbotapi.NewKeyboardButton("3"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("4"),
					tgbotapi.NewKeyboardButton("5"),
					tgbotapi.NewKeyboardButton("6"),
				))
		case "assignments":
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
					tgbotapi.NewInlineKeyboardButtonData("2", "2"),
					tgbotapi.NewInlineKeyboardButtonData("3", "3"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("4", "4"),
					tgbotapi.NewInlineKeyboardButtonData("5", "5"),
					tgbotapi.NewInlineKeyboardButtonData("6", "6"),
				),
			)
		case "close":
			msg.Text = "done"
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}
		if _, err := bot.Send(msg); err != nil {
			log.Error(err)
		}
	}
}
