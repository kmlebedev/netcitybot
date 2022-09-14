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
	btTypeLogin = "login"
)

func Login(msg *tgbotapi.MessageConfig, api *netcity.ClientApi) {
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	schoolTextSize := 0
	for _, schoolNames := range api.SchoolGroups {
		for i, schoolName := range schoolNames {
			schoolIdx := api.Schools[schoolName]
			schoolNameArr := strings.Split(schoolName, "№")
			if i != 0 && len(schoolNameArr) == 2 {
				schoolName = schoolNameArr[1]
			}
			if schoolTextSize > 50 {
				rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
				kbButRow = tgbotapi.NewInlineKeyboardRow()
				schoolTextSize = 0
			}
			kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(schoolName,
				fmt.Sprintf("%s:%d", btTypeLogin, schoolIdx)))
			schoolTextSize += len(schoolName)
		}
		rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
		kbButRow = tgbotapi.NewInlineKeyboardRow()
		schoolTextSize = 0
	}
	msg.ReplyMarkup = rpKeyboard
}

func GetUpdates(bot *tgbotapi.BotAPI, api *netcity.ClientApi) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		// Обработка Inline кнопок
		var msg tgbotapi.MessageConfig
		if update.CallbackQuery != nil && update.CallbackQuery.Data != "" && update.CallbackQuery.Message != nil {
			data := strings.Split(update.CallbackQuery.Data, ":")
			if len(data) < 2 && data[0] != "" {
				continue
			}
			switch data[0] { // Button Data Type
			case btTypeLogin:
				// log.Infof("update: %+v", update)
				if schoolId, err := strconv.Atoi(data[1]); err != nil {
					log.Warningf("%v: %+v", btTypeLogin, err)
					if schoolName, ok := api.SchoolIdsToName[int32(schoolId)]; ok {
						msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
							fmt.Sprintf("Введите ваш логин для %s", schoolName))
					} else {
						log.Warningf("%v: school id:%d not found", btTypeLogin, schoolId)
					}
				}
			default:
				continue
			}
		}
		// Todo убрать
		if update.Message != nil {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "start":
				msg.Text = "Выьерете вашу оранизацию"
				Login(&msg, api)
			case "hello":
				msg.Text = "И тебе привет."
			case "login":
				msg.Text = "login"
				Login(&msg, api)
			case "logout":
				msg.Text = "logout"
			}
			switch update.Message.Text {
			case "diary":
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("1"),
						tgbotapi.NewKeyboardButton("2"),
						tgbotapi.NewKeyboardButton("3"),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("4"),
						tgbotapi.NewKeyboardButton("5"),
						tgbotapi.NewKeyboardButton("6"),
					))
			case "assignments":
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
						tgbotapi.NewInlineKeyboardButtonData("2", "2"),
						tgbotapi.NewInlineKeyboardButtonData("3", "3"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("4", "4"),
						tgbotapi.NewInlineKeyboardButtonData("5", "5"),
						tgbotapi.NewInlineKeyboardButtonData("6", "6"),
					),
				)
			case "close":
				msg.Text = "done"
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			}
		}

		if msg.Text != "" {
			if _, err := bot.Send(msg); err != nil {
				log.Error(err)
			}
		}
	}
}
