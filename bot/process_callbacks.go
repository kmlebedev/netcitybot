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
func ProcessCallbackQuery(update tgbotapi.Update, sendMsg *tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	sendMsg.ChatID = update.CallbackQuery.Message.Chat.ID
	user := GetChatUser(update.Message.From.ID)
	sendMsg.Text = update.CallbackQuery.Data
	dataArr := strings.Split(update.CallbackQuery.Data, ":")
	switch dataArr[0] { // Button Data Type
	case BtTypeState: // state:name Нажатие на кнопку региона
		user.StateName = dataArr[1]
		ReplySelectProvince(sendMsg, dataArr[1])

	case BtTypeProvince: // province:name Нажатие на кнопку населённого пункту
		user.ProvinceName = dataArr[1]
		ReplySelectCity(sendMsg, dataArr[1])

	case BtTypeCity: // city:name Нажатие на кнопку города
		user.CityName = dataArr[1]
		ReplySelectSchool(sendMsg, dataArr[1])

	case BtTypeSchool: // school:id Нажатие на кномку школы
		if len(dataArr) != 3 {
			return
		}
		urlId, _ := strconv.Atoi(dataArr[1])
		schoolId, _ := strconv.Atoi(dataArr[2])
		if school := UrlSchools[uint64(urlId)][int32(schoolId)]; school != nil {
			user.School = school
			sendMsg.Text = fmt.Sprintf("%s %s", MsgReqLogin, school.Name)
		}

	case BtTypeStudent: // student:id
		netCityApi := GetLoginWebApi(update.Message.From.ID)
		if netCityApi == nil {
			sendMsg.Text = "Вы не вошли в дневник"
			return
		}
		if user.Assignments != nil {
			sendMsg.Text = "Пересылка заданий уже включена"
			return
		}
		if len(dataArr) != 2 {
			return
		}
		studentId, _ := strconv.Atoi(dataArr[1])
		go netCityApi.LoopPullingOrder(300, bot, sendMsg.ChatID, &[]int{studentId}, user)
		sendMsg.Text = "Включена пересылка новых заданий"

	default:
		log.Warningf("callback query data %+v not process", update.CallbackQuery.Data)
	}
}
