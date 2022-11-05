package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	"github.com/kmlebedev/netcitybot/bot/storage"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"time"
)

func ProcessText(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig) {
	switch updateMsg.Text {
	case "diary":
		sendMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("3"),
			))
	case "assignments":
		if api := GetLoginWebApi(updateMsg.Chat.ID); api != nil {
			currentTime := time.Now()
			weekStrat := currentTime.AddDate(0, 0, 0)
			weekEnd := currentTime.AddDate(0, 0, 8)
			assignments, err := api.GetAssignments(
				api.Uid,
				weekStrat.Format("2006-01-02"),
				weekEnd.Format("2006-01-02"),
				false,
				false,
				api.CurrentYearId,
			)
			if err != nil {
				sendMsg.Text = fmt.Sprintf("Что то пошло не так: %+v", err)
				log.Warningf("GetAssignments: %+v", err)
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
		}
	case "close":
		sendMsg.Text = "done"
		sendMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	default:
		if updateMsg.Chat.IsPrivate() {
			if login, ok := ChatUsers[updateMsg.Chat.ID]; ok {
				switch {

				// Обработываем ввод логина
				case updateMsg.MessageID == login.ReqNameMsgId+1:
					login.LoginName = updateMsg.Text
					login.ReqPasswdMsgId = updateMsg.MessageID
					sendMsg.Text = fmt.Sprintf("%s %s", MsgReqPasswd, login.LoginName)

				// Обработываем ввод пароля
				case updateMsg.MessageID == login.ReqPasswdMsgId+1:
					loginPassword := updateMsg.Text
					netcityConfig := netcity.Config{
						Url:      login.NetCityUrl,
						SchoolId: login.SchoolId,
						Username: login.LoginName,
						Password: loginPassword,
					}
					if netCityApi, err := netcity.NewClientApi(&netcityConfig); err == nil {
						login.NetCityApi = netCityApi
						sendMsg.Text = fmt.Sprintf("Данные верны")
						// Todo под учеткой родителя необходиямо явно передавать id класса
						if students, err := netCityApi.GetStudents(0); err == nil {
							sendMsg.Text += fmt.Sprintf(" и в вашем класса %d учеников", len(*students))
						}

					} else {
						sendMsg.Text = fmt.Sprintf("Данные не верны или повторите попытку позже: %+v", err)
						log.Warningf("BotLogin err: %+v", err)
					}
					// Сохраняем данные для логина
					ChatLogins.NewUserLoginData(updateMsg.Chat.ID, &storage.UserLoginData{
						NetCityUrl: netcityConfig.Url,
						CityId:     login.CityId,
						SchoolId:   netcityConfig.SchoolId,
						Login:      netcityConfig.Username,
						Password:   netcityConfig.Password,
						CityName:   login.CityName,
						SchoolName: login.SchoolName,
					})
				}
			}
		}
	}
}
