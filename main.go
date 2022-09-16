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
	EnvKeyNetCityUrls     = "NETCITY_URLS"        // https://netcity.eimc.ru,http//lync.schoolroo.ru"
	EnvKeyNetStudentIds   = "NETCITY_STUDENT_IDS" // 71111,75555
	EnvKeyYearId          = "NETCITY_YEAR_ID"
	EnvKeySyncEnabled     = "NETCITY_SYNC_ENABLED"
	EnvKeyRedisAddress    = "REDIS_ADDRESS"
	EnvKeyRedisDB         = "REDIS_DB"
	EnvKeyRedisPassword   = "REDIS_PASSWORD"
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

func main() {
	var netcityApi *netcity.ClientApi
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

	token := os.Getenv(EnvKeyTgbotToken)
	if token == "" {
		log.Fatalf("bot api token not found in env key: %s", EnvKeyTgbotToken)
	}

	chatId, err := strconv.ParseInt(os.Getenv(EnvKeyTgbotChatId), 10, 64)
	if err != nil {
		log.Warningf("bot chat id error in env key %s: %s", EnvKeyTgbotChatId, err)
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
	netCityUrl := TrimUrl(os.Getenv(EnvKeyNetCityUrl))
	if netCityUrl != "" {
		if netcityApi, err = netcity.NewClientApi(&netcity.Config{
			Url:      netCityUrl,
			School:   os.Getenv(EnvKeyNetCitySchool),
			Username: os.Getenv(EnvKeyNetCityUsername),
			Password: os.Getenv(EnvKeyNetCityPassword),
		}); err != nil {
			log.Warning(err)
		}

	}
	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	if netcityApi != nil {
		pullStudentIds := GetPullStudentIds()
		// sync assignments details with attachments to telegram
		if chatId > 0 && len(pullStudentIds) > 0 {
			assignments := map[int]netcity.DiaryAssignmentDetail{}
			go netcityApi.LoopPullingOrder(300, botApi, chatId, CurrentyYearId(netcityApi), rdb, &assignments, &pullStudentIds)
		}
		DoSync(netcityApi)
	}

	//botApi.Debug = true
	netCityUrls := []string{}
	singelUrlFound := false
	for _, url := range strings.Split(os.Getenv(EnvKeyNetCityUrls), ",") {
		if url == "" {
			continue
		}
		urlTrimed := TrimUrl(url)
		if urlTrimed == netCityUrl {
			singelUrlFound = true
		}
		netCityUrls = append(netCityUrls, urlTrimed)
	}
	if netCityUrl != "" && !singelUrlFound {
		netCityUrls = append(netCityUrls, netCityUrl)
	}
	bot.GetUpdates(botApi, netcityApi, &netCityUrls)
}
