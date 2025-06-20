package cache

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
	Rdb *redis.Client
)

func Init() error {
	addr := getEnv("REDIS_ADDR", "redis:6379")
	pass := os.Getenv("REDIS_PASS")        // 없으면 빈 문자열
	dbNum, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	Rdb = redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    pass,
		DB:          dbNum,
		PoolSize:    10,
		MinIdleConns: 5,
	})

	// 1초 내로 PING 응답 없으면 에러
	ctx, cancel := context.WithTimeout(Ctx, time.Second)
	defer cancel()
	return Rdb.Ping(ctx).Err()
}

func getEnv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}

