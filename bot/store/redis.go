package store

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kmlebedev/netcitybot/config"
	"os"
	"strconv"
	"time"
)

const (
	keyUrls = "urls"
	keyUser = "user"
)

type RedisStore struct {
	db  *redis.Client
	opt *redis.Options
}

var (
	ctx = context.Background()
)

func (s *RedisStore) New() (*RedisStore, error) {
	redisOpt := redis.Options{
		Addr:     os.Getenv(config.EnvKeyRedisAddress),
		Password: os.Getenv(config.EnvKeyRedisPassword),
	}
	if redisOpt.Addr == "" || redisOpt.Password == "" {
		return nil, fmt.Errorf("redis options Addr or Password is empty")
	}
	if db, err := strconv.Atoi(os.Getenv(config.EnvKeyRedisDB)); err == nil {
		redisOpt.DB = db
	}
	rdb := redis.NewClient(&redisOpt)
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis Ping: %+v", err)
	}
	return &RedisStore{db: rdb, opt: &redisOpt}, nil
}

func (s *RedisStore) getUrlIdx(newUrl string) (int32, bool) {
	if urls, err := s.db.LRange(ctx, keyUrls, 0, -1).Result(); err == nil {
		for i, url := range urls {
			if url == newUrl {
				return int32(i), true
			}
		}
	}
	return -1, false
}
