package storage

type UserLoginData struct {
	NetCityUrl string
	UserName   string
	Password   string
	CityName   string
	SchoolName string
	CityId     int32
	SchoolId   int
}

type StorageMap interface {
	GetNetCityUrls() map[uint64]string
	UpdateNetCityUrls(urls *[]string)
	GetUserLoginData(chatId int64) *UserLoginData
	UpdateUserLoginData(chatId int64, newUserLoginData UserLoginData)
	NewUserLoginData(chatId int64, userLoginData *UserLoginData)
	GetNetCityUrl(chatId int64) string
	GetUserName(chatId int64) string
	GetPassword(chatId int64) string
	GetCityName(chatId int64) string
	GetSchoolName(chatId int64) string
	GetCityId(chatId int64) int32
	GetSchoolId(chatId int64) int
	SetNetCityUrl(chatId int64, netCityUrl string)
	SetUserName(chatId int64, userName string)
	SetPassword(chatId int64, password string)
	SetCityName(chatId int64, cityName string)
	SetSchoolName(chatId int64, schoolName string)
	SetCityId(chatId int64, cityId int32)
	SetSchoolId(chatId int64, schoolId int)
}
