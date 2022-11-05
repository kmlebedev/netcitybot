package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

// Обработываем нажания кнопок
func ProcessCallbackQuery(update tgbotapi.Update, sendMsg *tgbotapi.MessageConfig) {
	sendMsg.ChatID = update.CallbackQuery.Message.Chat.ID
	sendMsg.Text = update.CallbackQuery.Data
	dataArr := strings.Split(update.CallbackQuery.Data, ":")
	switch dataArr[0] { // Button Data Type
	case BtTypeCity: // city:name Нажатие на кнопку города
		if _, ok := ChatUsers[sendMsg.ChatID]; ok {
			ChatUsers[sendMsg.ChatID].CityName = dataArr[1]
			ReplySelectSchool(sendMsg, dataArr[1])
		}

	case BtTypeSchool: // school:id Нажатие на кномку школы
		if len(dataArr) != 3 {
			return
		}
		urlId, _ := strconv.Atoi(dataArr[1])
		schoolId, _ := strconv.Atoi(dataArr[2])
		if _, ok := ChatUsers[sendMsg.ChatID]; ok {
			ChatUsers[sendMsg.ChatID].SchoolId = schoolId
			school := GetSchool(int32(urlId), int32(schoolId))
			// Todo data race
			if school != nil {
				ChatUsers[sendMsg.ChatID].CityId = school.Id
				ChatUsers[sendMsg.ChatID].NetCityUrl = NetCityUrls[school.UlrId]
				ChatUsers[sendMsg.ChatID].SchoolName = school.Name
				sendMsg.Text = fmt.Sprintf("%s %s", MsgReqLogin, school.Name)
			}
			//log.Warningf("%v: school id:%d not found", btTypeLogin, schoolId)
		}
	default:
		log.Warningf("callback query data %+v not process", update.CallbackQuery.Data)
	}
}
