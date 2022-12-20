package storageMemory

import (
	netcity_pb "github.com/kmlebedev/netcitybot/pb/netcity"
	"sync"
)

type StorageMem struct {
	NetCityUrls          map[uint64]string
	ChatToUserLoginData  map[int64]*netcity_pb.AuthParam
	ChatToUserConfigData map[int64]*netcity_pb.UserConfig
	lock                 sync.RWMutex
}

func NewStorageMem() *StorageMem {
	return &StorageMem{
		ChatToUserLoginData: make(map[int64]*netcity_pb.AuthParam),
		lock:                sync.RWMutex{},
	}
}
func (s *StorageMem) GetUserLoginData(chatId int64) *netcity_pb.AuthParam {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.ChatToUserLoginData != nil {
		if d, ok := s.ChatToUserLoginData[chatId]; ok {
			return d
		}
	}

	return nil
}

func (s *StorageMem) GetNetCityUrls() (urls map[uint64]string) {
	return s.NetCityUrls
}

func (s *StorageMem) UpdateNetCityUrls(urls *[]string) {
	if s.NetCityUrls == nil && len(*urls) != 0 {
		s.NetCityUrls = map[uint64]string{}
	}
	for i, url := range *urls {
		s.NetCityUrls[uint64(i)] = url
	}
}

func (s *StorageMem) PutUserLoginData(chatId int64, userLoginData *netcity_pb.AuthParam) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ChatToUserLoginData == nil {
		s.ChatToUserLoginData = make(map[int64]*netcity_pb.AuthParam)
	}
	s.ChatToUserLoginData[chatId] = userLoginData
}

func (s *StorageMem) UpdateUserLoginData(chatId int64, newUserLoginData *netcity_pb.AuthParam) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.ChatToUserLoginData == nil {
		s.ChatToUserLoginData = make(map[int64]*netcity_pb.AuthParam)
	}

	if d, ok := s.ChatToUserLoginData[chatId]; ok {
		if newUserLoginData.Cid != 0 {
			d.Cid = newUserLoginData.Cid
		}
		if newUserLoginData.Pid != 0 {
			d.Pid = newUserLoginData.Pid
		}
		if newUserLoginData.Sid != 0 {
			d.Sid = newUserLoginData.Sid
		}
		if newUserLoginData.Cn != 0 {
			d.Cn = newUserLoginData.Cn
		}
		if newUserLoginData.Scid != 0 {
			d.Scid = newUserLoginData.Scid
		}
		if newUserLoginData.UrlId != 0 {
			d.UrlId = newUserLoginData.UrlId
		}
		if newUserLoginData.UN != "" {
			d.UN = newUserLoginData.UN
		}
		if newUserLoginData.PW != "" {
			d.PW = newUserLoginData.PW
		}
	} else {
		s.ChatToUserLoginData[chatId] = newUserLoginData
	}
}

func (s *StorageMem) GetUserConfigData(chatId int64) *netcity_pb.UserConfig {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.ChatToUserConfigData != nil {
		if d, ok := s.ChatToUserConfigData[chatId]; ok {
			return d
		}
	}

	return nil
}

func (s *StorageMem) PutUserConfigData(chatId int64, userConfigData *netcity_pb.UserConfig) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ChatToUserConfigData == nil {
		s.ChatToUserConfigData = make(map[int64]*netcity_pb.UserConfig)
	}
	s.ChatToUserConfigData[chatId] = userConfigData
}
