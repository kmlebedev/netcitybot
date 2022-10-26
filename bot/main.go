package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	btTypeCity   = "city"
	btTypeSchool = "school"
	msgReqLogin  = "Введите ваш логин для"
	msgReqPasswd = "Введите ваш пароль для логина"
	btRowMaxSize = 40
)

type User struct {
	NetCityUrl     string
	Name           string
	Password       string
	CityName       string
	SchoolName     string
	CityId         int
	SchoolId       int
	ReqNameMsgId   int
	ReqPasswdMsgId int
	SentMsgLastId  int
	NetCityConfig  *netcity.Config
	NetCityApi     *netcity.ClientApi
	Marks          map[int]netcity.AssignmentMark
	TrackMarksCn   chan bool
	Valid          bool
}

var (
	Chatlogins = map[int64]*User{}
)

func GetLoginWebApi(chatId int64) *netcity.ClientApi {
	if _, ok := Chatlogins[chatId]; ok {
		return Chatlogins[chatId].NetCityApi
	}
	return nil
}

func GetSchool(urlId int32, id int32) *School {
	for _, school := range Schools {
		if school.UlrId == urlId && school.Id == id {
			return &school
		}
	}
	return nil
}

// Обработываем нажания кнопок
func ProcessCallbackQuery(update tgbotapi.Update, sendMsg *tgbotapi.MessageConfig) {
	sendMsg.ChatID = update.CallbackQuery.Message.Chat.ID
	sendMsg.Text = update.CallbackQuery.Data
	dataArr := strings.Split(update.CallbackQuery.Data, ":")
	switch dataArr[0] { // Button Data Type
	case btTypeCity: // city:name Нажатие на кнопку города
		if _, ok := Chatlogins[sendMsg.ChatID]; ok {
			Chatlogins[sendMsg.ChatID].CityName = dataArr[1]
			ReplySelectSchool(sendMsg, dataArr[1])
		}

	case btTypeSchool: // school:id Нажатие на кномку школы
		if len(dataArr) != 3 {
			return
		}
		urlId, _ := strconv.Atoi(dataArr[1])
		schoolId, _ := strconv.Atoi(dataArr[2])
		if _, ok := Chatlogins[sendMsg.ChatID]; ok {
			Chatlogins[sendMsg.ChatID].SchoolId = schoolId
			school := GetSchool(int32(urlId), int32(schoolId))
			// Todo data race
			if school != nil {
				Chatlogins[sendMsg.ChatID].NetCityUrl = NetCityUrls[school.UlrId]
				sendMsg.Text = fmt.Sprintf("%s %s", msgReqLogin, school.Name)
			}
			//log.Warningf("%v: school id:%d not found", btTypeLogin, schoolId)
		}
	default:
		log.Warningf("callback query data %+v not process", update.CallbackQuery.Data)
	}
}

func trackMarks(login *User) (string, error) {
	var msg string
	marks, err := login.NetCityApi.GetLessonAssignmentMarks()
	if err != nil {
		return msg, fmt.Errorf("Ошибка получения оценок: %+v", err)
	}
	if len(marks) == 0 {
		return msg, nil
	}
	if isEq := reflect.DeepEqual(login.Marks, marks); isEq {
		return msg, nil
	}

	for id, markNew := range marks {
		markOld, found := login.Marks[id]
		if found && reflect.DeepEqual(markNew, markOld) {
			continue
		}
		msg += fmt.Sprintf("%+v\n", markNew)
	}
	login.Marks = marks
	return msg, nil
}

func ProcessCommand(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	switch updateMsg.Command() {
	case "contacts":
		if login, ok := Chatlogins[sendMsg.ChatID]; ok && login.NetCityApi != nil {
			if mobilePhone, email, err := login.NetCityApi.GetContacts(); err == nil {
				sendMsg.Text = fmt.Sprintf("Мобильный телефон: %s и email: %s", mobilePhone, email)
			}
		} else {
			sendMsg.Text = "Вы не привязали дневник"
		}
	case "start":
		ReplySelectCity(sendMsg)
	case "track_marks":
		login, ok := Chatlogins[sendMsg.ChatID]
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
		if login, ok := Chatlogins[sendMsg.ChatID]; ok && login.NetCityApi != nil {
			login.NetCityApi.Logout()
		}
	}
}

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
			if login, ok := Chatlogins[updateMsg.Chat.ID]; ok {
				switch {
				// Обработываем ввод логина
				case updateMsg.MessageID == login.ReqNameMsgId+1:
					login.Name = updateMsg.Text
					login.ReqPasswdMsgId = updateMsg.MessageID
					sendMsg.Text = fmt.Sprintf("%s %s", msgReqPasswd, login.Name)
				// Обработываем ввод пароля
				case updateMsg.MessageID == login.ReqPasswdMsgId+1:
					login.Password = updateMsg.Text
					netcityConfig := netcity.Config{
						Url:      login.NetCityUrl,
						SchoolId: login.SchoolId,
						Username: login.Name,
						Password: login.Password,
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
				}
			}
		}
	}
}

func GetUpdates(bot *tgbotapi.BotAPI, api *netcity.ClientApi, urls *[]string) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	NetCityUrls = *urls
	prepareLoginData()

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		var msg tgbotapi.MessageConfig
		switch {
		// Обработка Inline кнопок
		case update.CallbackQuery != nil:
			ProcessCallbackQuery(update, &msg)
		// Обработка сообщений
		case update.Message != nil:
			msg.ChatID = update.Message.Chat.ID
			msg.Text = update.Message.Text
			switch {
			case update.Message.Command() != "":
				ProcessCommand(update.Message, &msg, bot)
			case update.Message.Text != "":
				ProcessText(update.Message, &msg)
				//log.Infof("UpdateID %+v: %+v,", update.UpdateID, update.Message)
			}
		}
		if msg.Text != "" {
			sentMsg, err := bot.Send(msg)
			if err != nil {
				log.Error(err)
			}
			if _, ok := Chatlogins[sentMsg.Chat.ID]; ok {
				Chatlogins[sentMsg.Chat.ID].SentMsgLastId = sentMsg.MessageID
				if strings.HasPrefix(sentMsg.Text, msgReqLogin) {
					Chatlogins[sentMsg.Chat.ID].ReqNameMsgId = sentMsg.MessageID
				} else if strings.HasPrefix(sentMsg.Text, msgReqPasswd) {
					Chatlogins[sentMsg.Chat.ID].ReqPasswdMsgId = sentMsg.MessageID
				}
			} else {
				Chatlogins[sentMsg.Chat.ID] = &User{SentMsgLastId: sentMsg.MessageID}
			}
		}
	}
}
