package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	. "github.com/kmlebedev/netcitybot/bot/constants"
	"github.com/kmlebedev/netcitybot/bot/storage"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
)

func ProcessTextPrivate(updateMsg *tgbotapi.Message, sendMsg *tgbotapi.MessageConfig, user *User, netcityApi *netcity.ClientApi) {
	switch {
	// Обработываем ввод логина
	case updateMsg.MessageID == user.ReqNameMsgId+1:
		user.LoginName = updateMsg.Text
		user.ReqPasswdMsgId = updateMsg.MessageID
		sendMsg.Text = fmt.Sprintf("%s %s", MsgReqPasswd, user.LoginName)

	// Обработываем ввод пароля
	case updateMsg.MessageID == user.ReqPasswdMsgId+1:
		loginPassword := updateMsg.Text
		netcityConfig := netcity.Config{
			Url:      user.NetCityUrl,
			SchoolId: user.SchoolId,
			Username: user.LoginName,
			Password: loginPassword,
		}
		if netCityApi, err := netcity.NewClientApi(&netcityConfig); err == nil {
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
		ChatNetCityDb.NewUserLoginData(updateMsg.Chat.ID, &storage.UserLoginData{
			NetCityUrl: netcityConfig.Url,
			SchoolId:   netcityConfig.SchoolId,
			UserName:   netcityConfig.Username,
			Password:   netcityConfig.Password,
			CityId:     user.CityId,
			CityName:   user.CityName,
			SchoolName: user.SchoolName,
		})
	}
}
