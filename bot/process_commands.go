package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"time"
)

func ProcessCommand(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	switch updateMsg.Command() {
	case "contacts":
		if login, ok := ChatUsers[sendMsg.ChatID]; ok && login.NetCityApi != nil {
			if mobilePhone, email, err := login.NetCityApi.GetContacts(); err == nil {
				sendMsg.Text = fmt.Sprintf("Мобильный телефон: %s и email: %s", mobilePhone, email)
			}
		} else {
			sendMsg.Text = "Вы не привязали дневник"
		}
	case "start":
		ReplySelectCity(sendMsg)
	case "track_marks":
		login, ok := ChatUsers[sendMsg.ChatID]
		if !ok || login.NetCityApi == nil {
			sendMsg.Text = fmt.Sprintf("Войдите в дневник")
			return
		}
		if login.Marks != nil {
			sendMsg.Text = ""
			if login.TrackMarksCn != nil {
				login.TrackMarksCn <- true
			}
			return
		}
		var err error
		login.Marks, err = login.NetCityApi.GetLessonAssignmentMarks()
		if err != nil {
			sendMsg.Text = fmt.Sprintf("Что то пошло не так: %+v", err)
			return
		}
		sendMsg.Text = fmt.Sprintf("Включено отслеживание отметок")
		login.TrackMarksCn = make(chan bool)
		go func(chatID int64, bot *tgbotapi.BotAPI, login *User) {
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
						if _, err = bot.Send(tgbotapi.NewMessage(chatID, msg)); err != nil {
							log.Warningf("bot.Send: %+v", err)
						}
					}
				}
			}
		}(sendMsg.ChatID, bot, login)

	case "hello":
		sendMsg.Text = "И тебе привет."
	case "login":
		sendMsg.Text = "login"
		ReplySelectCity(sendMsg)
	case "logout":
		sendMsg.Text = "logout"
		if login, ok := ChatUsers[sendMsg.ChatID]; ok && login.NetCityApi != nil {
			login.NetCityApi.Logout()
		}
	}
}
