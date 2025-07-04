package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx   = context.Background()
	Rdb   *redis.Client
)

const sessionDataTTL = time.Hour * 24 * 7
const sessionTokenTTL = time.Hour * 24 * 7

func Init() error {
	addr := getEnv("REDIS_ADDR", "redis:6379")
	pass := os.Getenv("REDIS_PASS") // 없으면 빈 문자열
	dbNum, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	Rdb = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DB:           dbNum,
		PoolSize:     10,
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

// parseJSONNumberToInt 헬퍼 함수 (이전 답변 참고 - 필요시 여기에 직접 포함하거나 별도 정의)
func parseJSONNumberToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		return int(v), nil
	case int64:
		return int(v), nil
	case string: // 직접 문자열로 온 숫자 (예: "279")
		// 값이 문자열이고, 내용이 "[숫자]" 형태라고 가정
		var arrResult json.Number // json.Number를 사용하면 다양한 숫자 표현에 유연
		errUnmarshal := json.Unmarshal([]byte(v), &arrResult)
		if errUnmarshal == nil && len(arrResult) > 0 {
			numVal, numErr := arrResult.Int64() // json.Number에서 Int64 추출
			if numErr != nil {
				return 0, fmt.Errorf("json.Number '%s'를 int64로 변환 실패: %w", v, numErr)
			}
			return int(numVal), nil
		} else {
			// "[숫자]" 형태가 아니거나 언마샬링 실패
			return 0, fmt.Errorf("json.Number '%s'를 int64로 변환 실패: %w", v, errUnmarshal)
		}

	case json.Number: // encoding/json으로 디코딩 시 json.Number로 올 수 있음
		i, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("json.Number '%s'를 int64로 변환 실패: %w", v.String(), err)
		}
		return int(i), nil
	default:
		return 0, fmt.Errorf("예상치 못한 숫자 타입: %T (값: %v)", value, value)
	}
}
