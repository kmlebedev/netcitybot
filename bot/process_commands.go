package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"time"
)

func Login(sendMsg *tgbotapi.MessageConfig, user *netcity.User) {
	if user.NetCityApi != nil {
		sendMsg.Text = "Вы уже вошли в дневник"
		return
	}
	if len(States) > 1 {
		ReplySelectState(sendMsg)
	} else if len(Provinces) > 1 {
		ReplySelectProvince(sendMsg, States[0].Name)
	} else if len(Provinces) == 1 {
		ReplySelectCity(sendMsg, Provinces[0].Name)
	}
}

func ProcessCommand(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	netCityApi := GetLoginWebApi(updateMsg.From.ID)
	if netCityApi == nil {
		if updateMsg.Command() == "subs_assignments" || updateMsg.Command() == "track_marks" || updateMsg.Command() == "get_contacts" || updateMsg.Command() == "get_marks" {
			sendMsg.Text = "Вы не вошли в дневник, выполните /login"
			return
		}
	}
	user := GetChatUser(updateMsg.From.ID)
	switch updateMsg.Command() {
	case "get_marks":
		marks, err := user.NetCityApi.GetLessonAssignmentMarks(([]int32{int32(user.NetCityApi.Uid)}), -1, 1)
		if err != nil {
			sendMsg.Text = fmt.Sprintf("Что то пошло не так: %+v", err)
			return
		}
		sendMsg.Text = ""
		for _, mark := range marks {
			sendMsg.Text += mark.Message(nil)
		}
		sendMsg.ParseMode = "markdown"

	case "get_contacts":
		if mobilePhone, email, err := netCityApi.GetContacts(); err == nil {
			sendMsg.Text = fmt.Sprintf("Мобильный телефон: %s и email: %s", mobilePhone, email)
		} else {
			log.Errorf("netCityApi.GetContacts: %v", err)
		}
	case "add_netcity_url":
		sendMsg.Text = "Отправьте http-адрес электронного дневники"

	case "start":
		Login(sendMsg, user)

	case "subs_assignments":
		if user.Assignments != nil {
			sendMsg.Text = ""
			if user.TrackAssignmentsCn != nil {
				user.TrackAssignmentsCn <- true
			}
			return
		}
		user.TrackAssignmentsCn = make(chan bool)
		if len(user.NetCityApi.DiaryInit.Students) == 1 {
			go netCityApi.LoopPullingOrder(300, bot, sendMsg.ChatID, &[]int{int(user.NetCityApi.DiaryInit.Students[0].StudentId)}, user)
			sendMsg.Text = "Включена пересылка новых заданий"
		} else if len(user.NetCityApi.DiaryInit.Students) > 1 {
			ReplySelectStudent(sendMsg, &user.NetCityApi.DiaryInit.Students)
		} else {
			sendMsg.Text = "Дневник не найден"
		}
	case "track_marks":
		if user.Marks != nil {
			sendMsg.Text = ""
			if user.TrackMarksCn != nil {
				user.TrackMarksCn <- true
			}
			return
		}

		var err error
		user.Marks, err = user.NetCityApi.GetLessonAssignmentMarks(user.NetCityApi.GetStudentsIds(), -14, 1)
		if err != nil {
			sendMsg.Text = fmt.Sprintf("Что то пошло не так: %+v", err)
			return
		}
		sendMsg.Text = fmt.Sprintf("Включено отслеживание отметок")
		user.TrackMarksCn = make(chan bool)
		go func(chatID int64, bot *tgbotapi.BotAPI, login *netcity.User) {
			tick := time.Tick(time.Duration(5) * time.Minute)
			for {
				select {
				case <-login.TrackMarksCn:
					login.Marks = nil
					if _, err := bot.Send(tgbotapi.NewMessage(chatID,
						fmt.Sprintf("Отключено отслеживание отметок"))); err != nil {
						log.Warningf("bot.Send: %+v", err)
					}
					return
				case <-tick:
					if msg, err := trackMarks(login); err == nil && msg != "" {
						tgMsg := tgbotapi.NewMessage(chatID, msg)
						tgMsg.DisableWebPagePreview = true
						tgMsg.ParseMode = "markdown"
						if _, err = bot.Send(tgMsg); err != nil {
							log.Warningf("bot.Send: %+v", err)
						}
					}
				}
			}
		}(sendMsg.ChatID, bot, user)

	case "help":
		ReplyHelp(sendMsg)
	case "login":
		Login(sendMsg, user)
	case "logout":
		netCityApi.Logout()
		sendMsg.Text = "logout"
	}
}
