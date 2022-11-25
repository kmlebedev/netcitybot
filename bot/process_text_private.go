package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	"github.com/kmlebedev/netcitybot/netcity"
	"github.com/kmlebedev/netcitybot/pb/netcity"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
)

func ProcessTextPrivate(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, user *netcity.User) {
	switch {
	// Обработываем ввод логина
	case updateMsg.MessageID == user.ReqNameMsgId+1:
		user.UserName = updateMsg.Text
		user.ReqPasswdMsgId = updateMsg.MessageID
		sendMsg.Text = fmt.Sprintf("%s %s", MsgReqPasswd, user.UserName)

	// Обработываем ввод пароля
	case updateMsg.MessageID == user.ReqPasswdMsgId+1:
		user.Password = updateMsg.Text
		if netCityApi, err := netcity.NewClientApi(NetCityUrls[user.School.UrlId], user.GetAuthParam()); err == nil {
			user.NetCityApi = netCityApi
			sendMsg.Text = fmt.Sprintf("Данные верны")
			// Todo под учеткой родителя необходиямо явно передавать id класса
			if students, err := netCityApi.GetStudents(0); err == nil {
				sendMsg.Text += fmt.Sprintf(" и в вашем класса %d учеников", len(*students))
			}

		} else {
			sendMsg.Text = fmt.Sprintf("Данные не верны или повторите попытку позже: %+v", err)
			log.Warningf("BotLogin err: %+v", err)
		}
		// Todo неоходимо запросить разрешение на сохранение даных на диск
		// Сохраняем данные для логина
		ChatNetCityDb.PutUserLoginData(updateMsg.Chat.ID, &netcity_pb.AuthParam{
			Sft:   user.School.Sft,
			Cid:   user.School.Country.Id,
			Scid:  user.School.Id,
			Pid:   user.School.Province.Id,
			Cn:    user.School.City.Id,
			Sid:   user.School.Id,
			UN:    user.UserName,
			PW:    user.Password,
			UrlId: uint32(user.School.UrlId),
		})
	default:
		if strings.HasPrefix(updateMsg.Text, "http:") || strings.HasPrefix(updateMsg.Text, "https:") {
			u, err := url.Parse(strings.ReplaceAll(updateMsg.Text, " ", ""))
			if err != nil || u.Host == "" {
				sendMsg.Text = fmt.Sprintf("Адрес электронного дневника не валидный: %v", err)
				return
			}
			baseUrl := fmt.Sprintf("%s://%s/", u.Scheme, u.Host)
			httpClient := http.DefaultClient
			loginData := GetLoginData(baseUrl, httpClient)
			if loginData == nil {
				sendMsg.Text = "Не удалось получить данные для входа в дневник"
				return
			}
			if !loginData.SchoolLogin {
				sendMsg.Text = "Вход под логинам школы не разрешон"
			}
			for _, urlVal := range ChatNetCityDb.GetNetCityUrls() {
				if urlVal == baseUrl {
					sendMsg.Text = "Адрес электронного дневника уже добавлен"
					return
				}
			}
			ChatNetCityDb.UpdateNetCityUrls(&([]string{baseUrl}))
			curSchoolsCount := len(Schools)
			for id, urlVal := range ChatNetCityDb.GetNetCityUrls() {
				if urlVal == baseUrl {
					GetPrepareLoginData(id, baseUrl, httpClient)
				}
			}
			sendMsg.Text = fmt.Sprintf("Вы добавили %d школ", len(Schools)-curSchoolsCount)
		}
	}
}
