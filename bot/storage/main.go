package storage

import netcity_pb "github.com/kmlebedev/netcitybot/pb/netcity"

type StorageMap interface {
	GetNetCityUrls() map[uint64]string
	UpdateNetCityUrls(urls *[]string)
	GetUserLoginData(chatId int64) *netcity_pb.AuthParam
	UpdateUserLoginData(chatId int64, newUserLoginData *netcity_pb.AuthParam)
	PutUserLoginData(chatId int64, userLoginData *netcity_pb.AuthParam)
	GetUserConfigData(chatId int64) *netcity_pb.UserConfig
	PutUserConfigData(chatId int64, userConfig *netcity_pb.UserConfig)
}
