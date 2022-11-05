package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/bot"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	"github.com/kmlebedev/netcitybot/bot/storage"
	storageMemory "github.com/kmlebedev/netcitybot/bot/storage/memory"
	_ "github.com/kmlebedev/netcitybot/bot/storage/redis"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	ChatLogins  storage.StorageMap
	netcityApi  *netcity.ClientApi
	netcityUrls []string
	botApi      *tgbotapi.BotAPI
	botApiToken string
	botChatId   int64
)

func IsSyncEnabled() bool {
	if b, err := strconv.ParseBool(os.Getenv(EnvKeySyncEnabled)); err == nil {
		return b
	}
	return false
}

func CurrentyYearId(netcityApi *netcity.ClientApi) int {
	if yearId, err := strconv.Atoi(os.Getenv(EnvKeyYearId)); err == nil {
		return yearId
	}
	if netcityApi.CurrentYearId != 0 {
		return netcityApi.CurrentYearId
	}
	if currentyYearId, err := netcityApi.GetCurrentyYearId(); err == nil {
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

func DoSync(netcityApi *netcity.ClientApi) {
	if !IsSyncEnabled() {
		return
	}
	if err := netcityApi.GetClasses(); err != nil {
		netcityApi.Logout()
		log.Fatalf("netcity get classes: %+v", err)
	}
	if err := netcityApi.GetAllStudents(); err != nil || len(netcityApi.Students) == 0 {
		netcityApi.Logout()
		log.Fatalf("netcity get all students:%+v %+v", len(netcityApi.Students), err)
	}
	log.Infof("Sync years: %d, classes: %d, students: %d",
		len(netcityApi.Years), len(netcityApi.Classes), len(netcityApi.Students))
}

func TrimUrl(url string) string {
	return strings.TrimRight(strings.Trim(url, " "), "/")
}

func init() {
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

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
		if netcityApi != nil {
			netcityApi.Logout()
		}
		log.Info(sig)
		done <- true
	}()

	botApiToken = os.Getenv(EnvKeyTgbotToken)
	if botApiToken == "" {
		log.Fatalf("bot api token not found in env key: %s", EnvKeyTgbotToken)
	}

	botChatId, err = strconv.ParseInt(os.Getenv(EnvKeyTgbotChatId), 10, 64)
	if err != nil {
		log.Warningf("bot chat id error in env key %s: %s", EnvKeyTgbotChatId, err)
	}
	ChatLogins = storageMemory.NewStorageMem()
	redisOpt := redis.Options{
		Addr:     os.Getenv(EnvKeyRedisAddress),
		Password: os.Getenv(EnvKeyRedisPassword),
	}
	if redisOpt.Password != "" {
		if db, err := strconv.Atoi(os.Getenv(EnvKeyRedisDB)); err != nil {
			redisOpt.DB = db
		}
		rdb := redis.NewClient(&redisOpt)
		if _, err := rdb.Ping(context.Background()).Result(); err != nil {
			log.Fatalf("Redis Db ping: %v", err)
		}
		//ChatLogins = storageRedis.NewStorageRdb(rdb)
	}

	netcityUrl := TrimUrl(os.Getenv(EnvKeyNetCityUrl))
	if netcityUrl != "" {
		if netcityApi, err = netcity.NewClientApi(&netcity.Config{
			Url:      netcityUrl,
			School:   os.Getenv(EnvKeyNetCitySchool),
			Username: os.Getenv(EnvKeyNetCityUsername),
			Password: os.Getenv(EnvKeyNetCityPassword),
		}); err != nil {
			log.Warning(err)
		}
	}

	singelUrlFound := false
	for _, url := range strings.Split(os.Getenv(EnvKeyNetCityUrls), ",") {
		if url == "" {
			continue
		}
		urlTrimed := TrimUrl(url)
		if urlTrimed == netcityUrl {
			singelUrlFound = true
		}
		netcityUrls = append(netcityUrls, urlTrimed)
	}
	if netcityUrl != "" && !singelUrlFound {
		netcityUrls = append(netcityUrls, netcityUrl)
	}

	botApi, err = tgbotapi.NewBotAPI(botApiToken)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	//botApi.Debug = true
	// Only sync assignments to telegram chat
	if netcityApi != nil {
		pullStudentIds := GetPullStudentIds()
		// sync assignments details with attachments to telegram
		if botChatId != 0 && len(pullStudentIds) > 0 {
			assignments := map[int]netcity.DiaryAssignmentDetail{}
			go netcityApi.LoopPullingOrder(300, botApi, botChatId, CurrentyYearId(netcityApi), &assignments, &pullStudentIds)
		}
		DoSync(netcityApi)
	}

	// Process message
	bot.GetUpdates(botApi, &netcityUrls, ChatLogins)
}
