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
	"sync"
)

var (
	ChatUsers     = make(map[int64]*netcity.User)
	ChatNetCityDb storage.StorageMap
	ChatUsersLock = sync.RWMutex{}
)

func GetChatUser(chatId int64) *netcity.User {
	ChatUsersLock.RLock()
	_, ok := ChatUsers[chatId]
	ChatUsersLock.RUnlock()
	if ok {
		return ChatUsers[chatId]
	}
	return NewChatUser(chatId)
}

func NewChatUser(chatId int64) *netcity.User {
	ChatUsersLock.Lock()
	defer ChatUsersLock.Unlock()
	ChatUsers[chatId] = &netcity.User{}
	return ChatUsers[chatId]
}

func GetLoginWebApi(chatId int64) *netcity.ClientApi {
	user := GetChatUser(chatId)
	if user == nil {
		return nil
	}
	if user.NetCityApi != nil {
		return user.NetCityApi
	}
	if userLoginData := ChatNetCityDb.GetUserLoginData(chatId); userLoginData != nil {
		clientApi, err := netcity.NewClientApi(NetCityUrls[uint64(userLoginData.UrlId)], userLoginData)
		if err != nil || clientApi == nil {
			log.Errorf("netcity.NewClientApi: %v", err)
			return nil
		}
		user.NetCityApi = clientApi
		return clientApi
	}
	return nil
}

func trackMarks(login *netcity.User) (string, error) {
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

func GetUpdates(bot *tgbotapi.BotAPI, chatNetCityDb storage.StorageMap) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	ChatNetCityDb = chatNetCityDb
	GetAllPrepareLoginData()

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		var msg tgbotapi.MessageConfig
		switch {
		// Обработка Inline кнопок
		case update.CallbackQuery != nil:
			ProcessCallbackQuery(update, &msg, bot)
		// Обработка сообщений
		case update.Message != nil:
			msg.ChatID = update.Message.Chat.ID
			msg.Text = update.Message.Text
			switch {
			case update.Message.Command() != "":
				ProcessCommand(update.Message, &msg, bot)
			case update.Message.Text != "":
				user := GetChatUser(update.Message.Chat.ID)
				netCityApi := GetLoginWebApi(update.Message.Chat.ID)
				if update.Message.Chat.IsPrivate() && netCityApi == nil {
					ProcessTextPrivate(update.Message, &msg, user)
				} else if netCityApi != nil {
					ProcessText(update.Message, &msg, user, netCityApi)
				} else {
					msg.Text = "Вы не вошли в дневник"
				}
			}
		}
		if msg.Text != "" {
			sentMsg, err := bot.Send(msg)
			if err != nil {
				log.Error(err)
			}
			if sentMsg.Chat == nil {
				return
			}
			if user := GetChatUser(sentMsg.Chat.ID); user != nil {
				user.SentMsgLastId = sentMsg.MessageID
				if strings.HasPrefix(sentMsg.Text, MsgReqLogin) {
					user.ReqNameMsgId = sentMsg.MessageID
				} else if strings.HasPrefix(sentMsg.Text, MsgReqPasswd) {
					user.ReqPasswdMsgId = sentMsg.MessageID
				}
			} else {
				newUser := NewChatUser(sentMsg.Chat.ID)
				newUser.SentMsgLastId = sentMsg.MessageID
			}
		}
	}
}
