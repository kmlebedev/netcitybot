package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"time"
)

func ProcessText(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, user *User, netcityApi *netcity.ClientApi) {
	switch updateMsg.Text {
	case "diary":
		sendMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("3"),
			))
	case "assignments":
		currentTime := time.Now()
		weekStrat := currentTime.AddDate(0, 0, 0)
		weekEnd := currentTime.AddDate(0, 0, 8)
		assignments, err := netcityApi.GetAssignments(
			netcityApi.Uid,
			weekStrat.Format("2006-01-02"),
			weekEnd.Format("2006-01-02"),
			false,
			false,
			netcityApi.CurrentYearId,
		)
		if err != nil {
			sendMsg.Text = fmt.Sprintf("Что то пошло не так: %+v", err)
			log.Errorf("netCityApi.GetAssignments: %+v", err)
		}
		sendMsg.Text = ""
		sendMsg.ParseMode = "markdown"
		sendMsg.DisableWebPagePreview = true
		for _, weekdays := range assignments.WeekDays {
			for _, lesson := range weekdays.Lessons {
				if len(lesson.Assignments) > 0 {
					sendMsg.Text += fmt.Sprintf("%s %s %s\n", lesson.DayString(), lesson.SubjectName, lesson.Assignments[0].AssignmentName)
				}
			}
		}

	case "close":
		sendMsg.Text = "done"
		sendMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	default:
		sendMsg.Text = updateMsg.Text
	}
}
