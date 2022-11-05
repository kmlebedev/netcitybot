package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	"github.com/kmlebedev/netcitybot/bot/storage"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type User struct {
	NetCityConfig  *netcity.Config
	NetCityApi     *netcity.ClientApi
	Marks          map[int]netcity.AssignmentMark
	NetCityUrl     string
	LoginName      string
	Password       string
	CityName       string
	SchoolName     string
	CityId         int32
	SchoolId       int
	SentMsgLastId  int
	ReqNameMsgId   int
	ReqPasswdMsgId int
	TrackMarksCn   chan bool
	Valid          bool
}

var (
	ChatUsers  = make(map[int64]*User)
	ChatLogins storage.StorageMap
)

func GetLoginWebApi(chatId int64) *netcity.ClientApi {
	if _, ok := ChatUsers[chatId]; ok {
		return ChatUsers[chatId].NetCityApi
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
		if found && markOld.Mark != markNew.Mark {
			msg += fmt.Sprintf("Оценка исправлена c %d на *%d* ", markOld.Mark, markNew.Mark)
		} else {
			msg += fmt.Sprintf("Оценка *%d* ", markNew.Mark)
		}
		msg += fmt.Sprintf("по предмету: %s, по теме: %s, за: %s\n", markNew.SubjectName, markNew.AssignmentName, markNew.Day.Format("02 Jan"))
	}
	login.Marks = marks
	return msg, nil
}

func GetUpdates(bot *tgbotapi.BotAPI, urls *[]string, chatLogins storage.StorageMap) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	NetCityUrls = *urls
	ChatLogins = chatLogins
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
			if _, ok := ChatUsers[sentMsg.Chat.ID]; ok {
				ChatUsers[sentMsg.Chat.ID].SentMsgLastId = sentMsg.MessageID
				if strings.HasPrefix(sentMsg.Text, MsgReqLogin) {
					ChatUsers[sentMsg.Chat.ID].ReqNameMsgId = sentMsg.MessageID
				} else if strings.HasPrefix(sentMsg.Text, MsgReqPasswd) {
					ChatUsers[sentMsg.Chat.ID].ReqPasswdMsgId = sentMsg.MessageID
				}
			} else {
				ChatUsers[sentMsg.Chat.ID] = &User{SentMsgLastId: sentMsg.MessageID}
			}
		}
	}
}
