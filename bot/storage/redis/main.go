package redis

import (
	"github.com/go-redis/redis/v8"
	. "github.com/kmlebedev/netcitybot/bot/storage"
)

type StorageRdb struct {
	rdb *redis.Client
}

func NewStorageRdb(rdb *redis.Client) *StorageRdb {
	return &StorageRdb{
		rdb: rdb,
	}
}

func (s *StorageRdb) GetUserLoginData(chatId int64) *UserLoginData {
	// not implemented
	return nil
}

func (s *StorageRdb) UpdateUserLoginData(chatId int64, newUserLoginData UserLoginData) {
	// not implemented
}

func (s *StorageRdb) NewUserLoginData(chatId int64, userLoginData *UserLoginData) {
	// not implemented
}

func (s *StorageRdb) GetNetCityUrl(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageRdb) GetUserName(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageRdb) GetPassword(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageRdb) GetCityName(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageRdb) GetSchoolName(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageRdb) GetCityId(chatId int64) int {
	// not implemented
	return 0
}

func (s *StorageRdb) GetSchoolId(chatId int64) int {
	// not implemented
	return 0
}

func (s *StorageRdb) SetNetCityUrl(chatId int64, netCityUrl string) {
	// not implemented
}

func (s *StorageRdb) SetUserName(chatId int64, userName string) {
	// not implemented
}

func (s *StorageRdb) SetPassword(chatId int64, password string) {
	// not implemented
}

func (s *StorageRdb) SetCityName(chatId int64, cityName string) {
	// not implemented
}

func (s *StorageRdb) SetSchoolName(chatId int64, schoolName string) {
	// not implemented
}

func (s *StorageRdb) SetCityId(chatId int64, cityId int) {
	// not implemented
}

func (s *StorageRdb) SetSchoolId(chatId int64, schoolId int) {
	// not implemented
}
