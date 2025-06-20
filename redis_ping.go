//go:build ignore
// +build ignore

// redis_ping.go
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()

	// ── 환경변수 읽기 ─────────────────────────
	addr := os.Getenv("REDIS_ADDR") // 예: "redis:6379"
	pass := os.Getenv("REDIS_PASS") // 비밀번호 없으면 ""
	dbNum := 0
	if s := os.Getenv("REDIS_DB"); s != "" {
		n, _ := strconv.Atoi(s)
		dbNum = n
	}

	// ── 클라이언트 생성 & PING ────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       dbNum,
	})
	must(rdb.Ping(ctx).Err())
	fmt.Printf("✅  Redis PING OK (%s)\n", addr)

	// ── SET & GET 테스트 ─────────────────────
	key := "test-key"
	val := fmt.Sprintf("time-%d", time.Now().Unix())

	must(rdb.Set(ctx, key, val, time.Minute).Err())
	got, err := rdb.Get(ctx, key).Result()
	must(err)

	if got == val {
		fmt.Printf("🎉  SET/GET 성공: %s = %s\n", key, got)
	} else {
		fmt.Printf("⚠️  값 불일치! 기대: %s, 실제: %s\n", val, got)
	}
}
