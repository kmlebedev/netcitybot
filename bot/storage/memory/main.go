package memory

import (
	. "github.com/kmlebedev/netcitybot/bot/storage"
	"sync"
)

type StorageMem struct {
	ChatToUserLoginData map[int64]*UserLoginData
	lock                sync.RWMutex
}

func NewStorageMem() *StorageMem {
	return &StorageMem{
		ChatToUserLoginData: make(map[int64]*UserLoginData),
		lock:                sync.RWMutex{},
	}
}
func (s *StorageMem) GetUserLoginData(chatId int64) *UserLoginData {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.ChatToUserLoginData != nil {
		if d, ok := s.ChatToUserLoginData[chatId]; ok {
			return d
		}
	}

	return nil
}

func (s *StorageMem) UpdateUserLoginData(chatId int64, newUserLoginData UserLoginData) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.ChatToUserLoginData == nil {
		s.ChatToUserLoginData = make(map[int64]*UserLoginData)
	}
	if s.ChatToUserLoginData != nil {
		if d, ok := s.ChatToUserLoginData[chatId]; ok {
			if newUserLoginData.NetCityUrl != "" {
				d.NetCityUrl = newUserLoginData.NetCityUrl
			}
			if newUserLoginData.Login != "" {
				d.Login = newUserLoginData.Login
			}
			if newUserLoginData.Password != "" {
				d.Password = newUserLoginData.Password
			}
			if newUserLoginData.CityName != "" {
				d.CityName = newUserLoginData.CityName
			}
			if newUserLoginData.SchoolName != "" {
				d.SchoolName = newUserLoginData.SchoolName
			}
			if newUserLoginData.CityId != 0 {
				d.CityId = newUserLoginData.CityId
			}
			if newUserLoginData.SchoolId != 0 {
				d.SchoolId = newUserLoginData.SchoolId
			}
		}
	}
}

func (s *StorageMem) NewUserLoginData(chatId int64, userLoginData *UserLoginData) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ChatToUserLoginData == nil {
		s.ChatToUserLoginData = make(map[int64]*UserLoginData)
	}
	s.ChatToUserLoginData[chatId] = userLoginData
}

func (s *StorageMem) GetNetCityUrl(chatId int64) string {
	return s.GetUserLoginData(chatId).NetCityUrl
}

func (s *StorageMem) GetName(chatId int64) string {
	return s.GetUserLoginData(chatId).Login
}

func (s *StorageMem) GetPassword(chatId int64) string {
	return s.GetUserLoginData(chatId).Password
}

func (s *StorageMem) GetCityName(chatId int64) string {
	return s.GetUserLoginData(chatId).CityName
}

func (s *StorageMem) GetSchoolName(chatId int64) string {
	return s.GetUserLoginData(chatId).SchoolName
}

func (s *StorageMem) GetCityId(chatId int64) int32 {
	return s.GetUserLoginData(chatId).CityId
}

func (s *StorageMem) GetSchoolId(chatId int64) int {
	return s.ChatToUserLoginData[chatId].SchoolId
}

func (s *StorageMem) SetNetCityUrl(chatId int64, netCityUrl string) {
	s.UpdateUserLoginData(chatId, UserLoginData{NetCityUrl: netCityUrl})
}

func (s *StorageMem) SetName(chatId int64, login string) {
	s.UpdateUserLoginData(chatId, UserLoginData{Login: login})
}

func (s *StorageMem) SetPassword(chatId int64, password string) {
	s.UpdateUserLoginData(chatId, UserLoginData{Password: password})
}

func (s *StorageMem) SetCityName(chatId int64, cityName string) {
	s.UpdateUserLoginData(chatId, UserLoginData{CityName: cityName})
}

func (s *StorageMem) SetSchoolName(chatId int64, schoolName string) {
	s.UpdateUserLoginData(chatId, UserLoginData{SchoolName: schoolName})
}

func (s *StorageMem) SetCityId(chatId int64, cityId int32) {
	s.UpdateUserLoginData(chatId, UserLoginData{CityId: cityId})
}

func (s *StorageMem) SetSchoolId(chatId int64, schoolId int) {
	s.UpdateUserLoginData(chatId, UserLoginData{SchoolId: schoolId})
}
