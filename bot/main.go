package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
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
	Valid          bool
}

var (
	Chatlogins = map[int64]*User{}
)

func GetSchool(id int32) *School {
	for _, school := range Schools {
		if school.Id == id {
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
		schoolId, _ := strconv.Atoi(dataArr[1])
		if _, ok := Chatlogins[sendMsg.ChatID]; ok {
			Chatlogins[sendMsg.ChatID].SchoolId = schoolId
			school := GetSchool(int32(schoolId))
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

func ProcessCommand(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, api *netcity.ClientApi) {
	switch updateMsg.Command() {
	case "start":
		ReplySelectCity(sendMsg)
	case "hello":
		sendMsg.Text = "И тебе привет."
	case "login":
		sendMsg.Text = "login"
		ReplySelectCity(sendMsg)
	case "logout":
		sendMsg.Text = "logout"
	}
}

func ProcessText(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, api *netcity.ClientApi) {
	switch updateMsg.Text {
	case "diary":
		sendMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("3"),
			))
	case "assignments":
		sendMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
				tgbotapi.NewInlineKeyboardButtonData("2", "2"),
				tgbotapi.NewInlineKeyboardButtonData("3", "3"),
			),
		)
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
						if students, err := netCityApi.GetStudents(0); err == nil {
							sendMsg.Text += fmt.Sprintf(" и в вашем класса %d учеников", len(*students))
						}
					} else {
						//sendMsg.Text = fmt.Sprintf("Данные не верны или повторите попытку позже")
						log.Warningf("BotLogin err: %+v", err)
						sendMsg.Text = err.Error()
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
				ProcessCommand(update.Message, &msg, api)
			case update.Message.Text != "":
				ProcessText(update.Message, &msg, api)
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
