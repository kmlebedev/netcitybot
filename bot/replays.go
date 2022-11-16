package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	swagger "github.com/kmlebedev/netSchoolWebApi/go"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
)

func ReplyHelp(msg *tgbotapi.MessageConfig) {
	msg.Text = "После входа в электронный дневник /start доступны команды:\n" +
		"/track_marks - подписка на новый оценки\n" +
		"/subs_assignments - пересылка домашних заданий в канал\n"
}

func ReplySelectStudent(msg *tgbotapi.MessageConfig, students *[]swagger.StudentDiaryInitStudents) {
	msg.Text = "Выберите ученика"
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0
	for _, student := range *students {
		if textSize > BtRowMaxSize {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(student.NickName,
			fmt.Sprintf("%s:%d", BtTypeStudent, student.StudentId)))
		textSize += len(student.NickName)
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	msg.ReplyMarkup = rpKeyboard
}

func ReplySelectState(msg *tgbotapi.MessageConfig) {
	msg.Text = "Выберите ваш регион"
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0

	stateMap := make(map[string]bool)
	stateNames := make([]string, 0, len(States))
	for _, c := range States {
		if _, ok := stateMap[c.Name]; !ok {
			stateMap[c.Name] = true
			stateNames = append(stateNames, c.Name)
		}
	}

	sort.Strings(stateNames)
	for _, stateName := range stateNames {
		if textSize > BtRowMaxSize {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(stateName,
			fmt.Sprintf("%s:%s", BtTypeState, stateName)))
		textSize += len(stateName)
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	msg.ReplyMarkup = rpKeyboard
}

func ReplySelectProvince(msg *tgbotapi.MessageConfig, stateName string) {
	msg.Text = "Выберите ваш район"
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0

	provinceMap := make(map[string]bool)
	provinceNames := make([]string, 0, len(Provinces))
	for _, p := range Provinces {
		if stateName != "" && p.State.Name != stateName {
			continue
		}
		if _, ok := provinceMap[p.Name]; !ok {
			provinceMap[p.Name] = true
			provinceNames = append(provinceNames, p.Name)
		}
	}
	if len(provinceNames) == 1 {
		ReplySelectCity(msg, provinceNames[0])
		return
	}

	sort.Strings(provinceNames)
	for _, provinceName := range provinceNames {
		if textSize > BtRowMaxSize {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		provinceNameBt := strings.Replace(provinceName, "район", "", 1)
		provinceNameBt = strings.Replace(provinceNameBt, "Городской округ", "", 1)
		provinceNameBt = strings.Trim(provinceNameBt, " ")
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(provinceNameBt,
			fmt.Sprintf("%s:%s", BtTypeProvince, provinceName)))
		textSize += len(provinceNameBt)
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	msg.ReplyMarkup = rpKeyboard
}

// Формируем кнопки Городов
func ReplySelectCity(msg *tgbotapi.MessageConfig, provinceName string) {
	msg.Text = "Выберите ваш населённый пункт"
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0

	cityMap := make(map[string]bool)
	cityNames := make([]string, 0, len(Cities))
	for _, c := range Cities {
		if provinceName != "" && c.Province.Name != provinceName {
			continue
		}
		if _, ok := cityMap[c.Name]; !ok {
			cityMap[c.Name] = true
			cityNames = append(cityNames, c.Name)
		}
	}
	if len(cityNames) == 1 {
		ReplySelectSchool(msg, cityNames[0])
		return
	}
	sort.Strings(cityNames)
	for _, cityName := range cityNames {
		if textSize > BtRowMaxSize {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(cityName,
			fmt.Sprintf("%s:%s", BtTypeCity, cityName)))
		textSize += len(cityName)
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	msg.ReplyMarkup = rpKeyboard
}

// Формируем кнопки Школ
func ReplySelectSchool(msg *tgbotapi.MessageConfig, cityName string) {
	msg.Text = fmt.Sprintf("Выберите ваш номер школы в населённом пункте %s", cityName)
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0
	var schoolsWithoutNum []int

	for i, school := range Schools {
		if cityName != "" && school.City.Name != cityName {
			continue
		}
		if textSize > BtRowMaxSize || len(kbButRow) >= 8 {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			//log.Infof("append to %s buttons: %d", cityName, len(kbButRow))
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		var schoolNum string
		if school.Num > 0 {
			schoolNum = fmt.Sprintf("%d", school.Num)
			kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(schoolNum,
				fmt.Sprintf("%s:%d:%d", BtTypeSchool, school.UrlId, school.Id)))
			textSize += len(schoolNum)
			log.Debugf("butten urlId %d, num: %s, id: %d, name %s", school.UrlId, schoolNum, school.Id, school.Name)
		} else {
			schoolsWithoutNum = append(schoolsWithoutNum, i)
		}
	}
	for _, i := range schoolsWithoutNum {
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(Schools[i].Name,
			fmt.Sprintf("%s:%d:%d", BtTypeSchool, Schools[i].UrlId, Schools[i].Id)))
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	//log.Infof("append to %s button rows: %d", cityName, len(rpKeyboard.InlineKeyboard))
	kbButRow = tgbotapi.NewInlineKeyboardRow()
	msg.ReplyMarkup = rpKeyboard
}
