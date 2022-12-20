package storageRedis

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	netcity_pb "github.com/kmlebedev/netcitybot/pb/netcity"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var ctx = context.Background()

const userLoginDataKey = "userLoginData"
const userConfigDataKey = "userConfigData"
const netCityUrlsKey = "netCityUrls"

func getUserLoginDataKey(chatId int64) string {
	return fmt.Sprintf("%s:%d", userLoginDataKey, chatId)
}

func getUserConfigDataKey(chatId int64) string {
	return fmt.Sprintf("%s:%d", userConfigDataKey, chatId)
}

func mtoFields(m protoreflect.Message) (fields []string) {
	m.Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		fields = append(fields, descriptor.TextName())
		return true
	})
	return
}

func mtoValues(m protoreflect.Message) (values []string) {
	m.Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		values = append(values, descriptor.TextName())
		values = append(values, value.String())
		return true
	})
	return
}

type StorageRdb struct {
	rdb *redis.Client
}

func NewStorageRdb(rdb *redis.Client) *StorageRdb {
	return &StorageRdb{
		rdb: rdb,
	}
}

func (s *StorageRdb) GetNetCityUrls() (urls map[uint64]string) {
	urlsArr, err := s.rdb.LRange(ctx, netCityUrlsKey, 0, -1).Result()
	if err != nil {
		log.Errorf("rdb.LRange netCityUrls %+v", err)
		return nil
	}
	urls = make(map[uint64]string)
	for i, url := range urlsArr {
		urls[uint64(i)] = url
	}
	return urls
}

func (s *StorageRdb) UpdateNetCityUrls(urls *[]string) {
	if len(*urls) == 0 {
		return
	}
	urlsArr, _ := s.rdb.LRange(ctx, netCityUrlsKey, 0, -1).Result()
	savedUrls := []string{}
	for _, url := range *urls {
		if !slices.Contains(urlsArr, url) {
			savedUrls = append(savedUrls, url)
		}
	}
	if len(savedUrls) == 0 {
		return
	}
	if err := s.rdb.RPush(ctx, netCityUrlsKey, savedUrls).Err(); err != nil {
		log.Errorf("rdb.RPush netCityUrls %+v", err)
	}
}

func (s *StorageRdb) GetUserLoginDataNativ(chatId int64) *netcity_pb.AuthParam {
	userLoginData := netcity_pb.AuthParam{}
	if vals, err := s.rdb.HMGet(ctx, getUserLoginDataKey(chatId), mtoFields(userLoginData.ProtoReflect())...).Result(); err == nil {
		for _, val := range vals {
			log.Debugf("HMGet AuthParam %+v", val)
		}
	}
	return nil
}

func (s *StorageRdb) PutUserLoginDataNativ(chatId int64, userLoginData *netcity_pb.AuthParam) {
	s.rdb.HSet(ctx,
		getUserLoginDataKey(chatId),
		mtoValues(userLoginData.ProtoReflect()),
	)
}

func (s *StorageRdb) UpdateUserLoginData(chatId int64, newUserLoginData *netcity_pb.AuthParam) {
	// not implemented
}

func (s *StorageRdb) PutUserLoginData(chatId int64, userLoginData *netcity_pb.AuthParam) {
	if value, err := proto.Marshal(userLoginData); err == nil {
		err := s.rdb.Set(ctx, getUserLoginDataKey(chatId), base64.StdEncoding.EncodeToString(value), 0).Err()
		if err != nil {
			log.Errorf("rdb.Set userLoginData %+v", err)
		}
	} else {
		log.Errorf("proto.Marshal userLoginData %+v", err)
	}
}

func (s *StorageRdb) GetUserLoginData(chatId int64) *netcity_pb.AuthParam {
	userLoginData := netcity_pb.AuthParam{}
	valueStr, err := s.rdb.Get(ctx, getUserLoginDataKey(chatId)).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("GetUserLoginData: s.rdb.Get userLoginData %+v", err)
		}
		return nil
	}

	value, err := base64.StdEncoding.DecodeString(valueStr)
	if err != nil {
		log.Errorf("GetUserLoginData: DecodeString userLoginData %+v", err)
		return nil
	}

	if err := proto.Unmarshal(value, &userLoginData); err != nil {
		log.Errorf("GetUserLoginData: proto.Unmarshal userLoginData %+v", err)
		return nil
	}

	return &userLoginData
}

func (s *StorageRdb) GetUserConfigData(chatId int64) *netcity_pb.UserConfig {
	userConfigData := netcity_pb.UserConfig{}
	valueStr, err := s.rdb.Get(ctx, getUserConfigDataKey(chatId)).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Errorf("GetUserConfig: s.rdb.Get userLoginData %+v", err)
		}
		return nil
	}

	value, err := base64.StdEncoding.DecodeString(valueStr)
	if err != nil {
		log.Errorf("GetUserConfig: DecodeString userConfigData %+v", err)
		return nil
	}

	if err := proto.Unmarshal(value, &userConfigData); err != nil {
		log.Errorf("GetUserConfig: proto.Unmarshal userConfigData %+v", err)
		return nil
	}

	return &userConfigData
}

func (s *StorageRdb) PutUserConfigData(chatId int64, userConfigData *netcity_pb.UserConfig) {
	if value, err := proto.Marshal(userConfigData); err == nil {
		err := s.rdb.Set(ctx, getUserConfigDataKey(chatId), base64.StdEncoding.EncodeToString(value), 0).Err()
		if err != nil {
			log.Errorf("rdb.Set userConfigData %+v", err)
		}
	} else {
		log.Errorf("proto.Marshal userConfigData %+v", err)
	}

}
