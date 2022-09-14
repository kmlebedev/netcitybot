package main

import (
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/bot"
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
	EnvKeySyncEnabled     = "NETCITY_SYNC_ENABLED"
	EnvKeyRedisAddress    = "REDIS_ADDRESS"
	EnvKeyRedisDB         = "REDIS_DB"
	EnvKeyRedisPassword   = "REDIS_PASSWORD"
)

var api *netcity.ClientApi

func IsSyncEnabled() bool {
	if b, err := strconv.ParseBool(os.Getenv(EnvKeySyncEnabled)); err == nil {
		return b
	}
	return false
}

func CurrentyYearId() int {
	if yearId, err := strconv.Atoi(os.Getenv(EnvKeyYearId)); err == nil {
		return yearId
	}
	if currentyYearId, err := api.GetCurrentyYearId(); err == nil {
		return currentyYearId
	} else {
		log.Fatalf("netcity year id error: %+v", err)
	}
	return 0
}

func GetPullStudentIds() (pullStudentIds []int) {
	for _, strId := range strings.Split(strings.TrimSpace(os.Getenv(EnvKeyNetStudentIds)), ",") {
		if id, err := strconv.Atoi(strings.Trim(strId, " ")); err == nil {
			pullStudentIds = append(pullStudentIds, id)
		}
	}
	return
}

func main() {
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

	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	// bot.Debug = true
	if IsSyncEnabled() {
		if err := api.GetClasses(); err != nil {
			api.Logout()
			log.Fatalf("netcity get classes: %+v", err)
		}
		if err := api.GetAllStudents(); err != nil || len(api.Students) == 0 {
			api.Logout()
			log.Fatalf("netcity get all students:%+v %+v", len(api.Students), err)
		}
		log.Infof("Sync years: %d, classes: %d, students: %d", len(api.Years), len(api.Classes), len(api.Students))
	}

	// sycn assignments details with attachments to telegram
	pullStudentIds := GetPullStudentIds()
	if chatId > 0 && len(pullStudentIds) > 0 {
		assignments := map[int]netcity.DiaryAssignmentDetail{}
		go api.LoopPullingOrder(60, botApi, chatId, CurrentyYearId(), rdb, &assignments, &pullStudentIds)
	}
	bot.GetUpdates(botApi, api)
}
