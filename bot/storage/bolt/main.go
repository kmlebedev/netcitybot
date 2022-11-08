package storageBolt

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	netcity_pb "github.com/kmlebedev/netcitybot/pb/netcity"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/exp/slices"
)

var (
	configBucket = []byte("netCityConfig")
	urlsBucket   = []byte("netCityUrls")
)

// itob returns an 8-byte big endian representation of v.
func uitob(v uint64) []byte {
	b := make([]byte, 8)
	binary.PutUvarint(b, v)
	return b
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, v)
	return b
}

type StorageBolt struct {
	db           *bolt.DB
	configBucket *bolt.Bucket
}

func NewStorageBolt(db *bolt.DB) *StorageBolt {
	return &StorageBolt{
		db: db,
	}
}

func (s *StorageBolt) GetNetCityUrls() (urls map[uint64]string) {
	s.db.View(func(tx *bolt.Tx) error {
		tx.Bucket(urlsBucket).ForEach(func(k, v []byte) error {
			key, _ := binary.Uvarint(k)
			urls[key] = string(v)
			return nil
		})
		return nil
	})
	return urls
}

func (s *StorageBolt) UpdateNetCityUrls(urls *[]string) {
	savedUrls := []string{}
	for _, url := range s.GetNetCityUrls() {
		if slices.Contains(*urls, url) {
			savedUrls = append(savedUrls, url)
		}
	}
	for _, url := range *urls {
		if slices.Contains(savedUrls, url) {
			continue
		}
		s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(urlsBucket)
			id, _ := b.NextSequence()
			return b.Put(uitob(id), []byte(url))
		})
	}
}

func (s *StorageBolt) GetUserLoginData(chatId int64) *netcity_pb.AuthParam {
	var authParam netcity_pb.AuthParam
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(configBucket)
		value := b.Get(itob(chatId))
		return proto.Unmarshal(value, &authParam)
	})
	return &authParam
}

func (s *StorageBolt) UpdateUserLoginData(chatId int64, newUserLoginData *netcity_pb.AuthParam) {
	value := s.GetUserLoginData(chatId)
	if value == nil {
		s.PutUserLoginData(chatId, newUserLoginData)
		return
	}
	if newUserLoginData.Cid != 0 {
		value.Cid = newUserLoginData.Cid
	}
	if newUserLoginData.Pid != 0 {
		value.Pid = newUserLoginData.Pid
	}
	if newUserLoginData.Sid != 0 {
		value.Sid = newUserLoginData.Sid
	}
	if newUserLoginData.Cn != 0 {
		value.Cn = newUserLoginData.Cn
	}
	if newUserLoginData.Scid != 0 {
		value.Scid = newUserLoginData.Scid
	}
	if newUserLoginData.UrlId != 0 {
		value.UrlId = newUserLoginData.UrlId
	}
	if newUserLoginData.UN != "" {
		value.UN = newUserLoginData.UN
	}
	if newUserLoginData.PW != "" {
		value.PW = newUserLoginData.PW
	}
	s.PutUserLoginData(chatId, value)
}

func (s *StorageBolt) PutUserLoginData(chatId int64, userLoginData *netcity_pb.AuthParam) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		value, err := proto.Marshal(userLoginData)
		if err != nil {
			return err
		}
		b := tx.Bucket(configBucket)
		return b.Put(itob(chatId), value)
	})
	if err != nil {
		log.Errorf("NewUserLoginData: %v", err)
	}
}
