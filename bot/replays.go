package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"sort"
)

// Формируем кнопки Городов
func ReplySelectCity(msg *tgbotapi.MessageConfig) {
	msg.Text = "Выберите ваш город"
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0

	cityMap := make(map[string]bool)
	cityNames := make([]string, 0, len(Cities))
	for _, c := range Cities {
		if _, ok := cityMap[c.Name]; !ok {
			cityMap[c.Name] = true
			cityNames = append(cityNames, c.Name)
		}
	}
	sort.Strings(cityNames)
	for _, cityName := range cityNames {
		if textSize > btRowMaxSize {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(cityName,
			fmt.Sprintf("%s:%s", btTypeCity, cityName)))
		textSize += len(cityName)
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	msg.ReplyMarkup = rpKeyboard
}

// Формируем кнопки Школ
func ReplySelectSchool(msg *tgbotapi.MessageConfig, cityName string) {
	msg.Text = "Выберите ваш номер школы"
	rpKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	kbButRow := tgbotapi.NewInlineKeyboardRow()
	textSize := 0
	var schoolsWithoutNum []int
	for i, school := range Schools {
		if school.City != cityName {
			continue
		}
		if textSize > btRowMaxSize || len(kbButRow) >= 8 {
			rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
			//log.Infof("append to %s buttons: %d", cityName, len(kbButRow))
			kbButRow = tgbotapi.NewInlineKeyboardRow()
			textSize = 0
		}
		var schoolNum string
		if school.Num > 0 {
			schoolNum = fmt.Sprintf("%d", school.Num)
			kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(schoolNum,
				fmt.Sprintf("%s:%d:%d", btTypeSchool, school.UlrId, school.Id)))
			textSize += len(schoolNum)
			log.Debugf("butten urlId %d, num: %s, id: %d, name %s", school.UlrId, schoolNum, school.Id, school.Name)
		} else {
			schoolsWithoutNum = append(schoolsWithoutNum, i)
		}
	}
	for _, i := range schoolsWithoutNum {
		kbButRow = append(kbButRow, tgbotapi.NewInlineKeyboardButtonData(Schools[i].Name,
			fmt.Sprintf("%s:%d:%d", btTypeSchool, Schools[i].UlrId, Schools[i].Id)))
	}
	rpKeyboard.InlineKeyboard = append(rpKeyboard.InlineKeyboard, kbButRow)
	//log.Infof("append to %s button rows: %d", cityName, len(rpKeyboard.InlineKeyboard))
	kbButRow = tgbotapi.NewInlineKeyboardRow()
	msg.ReplyMarkup = rpKeyboard
}
