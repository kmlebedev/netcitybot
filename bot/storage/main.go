package storage

type UserLoginData struct {
	NetCityUrl string
	Login      string
	Password   string
	CityName   string
	SchoolName string
	CityId     int32
	SchoolId   int
}

type StorageMap interface {
	GetUserLoginData(chatId int64) *UserLoginData
	UpdateUserLoginData(chatId int64, newUserLoginData UserLoginData)
	NewUserLoginData(chatId int64, userLoginData *UserLoginData)
	GetNetCityUrl(chatId int64) string
	GetName(chatId int64) string
	GetPassword(chatId int64) string
	GetCityName(chatId int64) string
	GetSchoolName(chatId int64) string
	GetCityId(chatId int64) int32
	GetSchoolId(chatId int64) int
	SetNetCityUrl(chatId int64, netCityUrl string)
	SetName(chatId int64, login string)
	SetPassword(chatId int64, password string)
	SetCityName(chatId int64, cityName string)
	SetSchoolName(chatId int64, schoolName string)
	SetCityId(chatId int64, cityId int32)
	SetSchoolId(chatId int64, schoolId int)
}
