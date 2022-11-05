package storageBolt

import (
	"encoding/binary"
	. "github.com/kmlebedev/netcitybot/bot/storage"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/exp/slices"
)

var (
	configBucket = []byte("netCityConfig")
	urlsBucket   = []byte("netCityUrls")
)

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
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
			urls[binary.BigEndian.Uint64(k)] = string(v)
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
			return b.Put(itob(id), []byte(url))
		})
	}
}

func (s *StorageBolt) GetUserLoginData(chatId int64) *UserLoginData {
	// not implemented
	return nil
}

func (s *StorageBolt) UpdateUserLoginData(chatId int64, newUserLoginData UserLoginData) {
	// not implemented
}

func (s *StorageBolt) NewUserLoginData(chatId int64, userLoginData *UserLoginData) {
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(configBucket)
		err := b.Put([]byte("answer"), []byte("42"))
		return err
	})
}

func (s *StorageBolt) GetNetCityUrl(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageBolt) GetUserName(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageBolt) GetPassword(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageBolt) GetCityName(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageBolt) GetSchoolName(chatId int64) string {
	// not implemented
	return ""
}

func (s *StorageBolt) GetCityId(chatId int64) int32 {
	// not implemented
	return 0
}

func (s *StorageBolt) GetSchoolId(chatId int64) int {
	// not implemented
	return 0
}

func (s *StorageBolt) SetNetCityUrl(chatId int64, netCityUrl string) {
	// not implemented
}

func (s *StorageBolt) SetUserName(chatId int64, userName string) {
	// not implemented
}

func (s *StorageBolt) SetPassword(chatId int64, password string) {
	// not implemented
}

func (s *StorageBolt) SetCityName(chatId int64, cityName string) {
	// not implemented
}

func (s *StorageBolt) SetSchoolName(chatId int64, schoolName string) {
	// not implemented
}

func (s *StorageBolt) SetCityId(chatId int64, cityId int32) {
	// not implemented
}

func (s *StorageBolt) SetSchoolId(chatId int64, schoolId int) {
	// not implemented
}
