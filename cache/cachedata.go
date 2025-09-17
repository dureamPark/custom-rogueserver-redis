package cache

import (
	"context"
	"encoding/base64"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"
)

// uuid로 한 유저의 cachedata가 있는지 확인
func IsValidCacheData(ctx context.Context, uuid []byte) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	exists, err := Rdb.Exists(ctx, redisKey).Result()

	// 데이터가 없다면?
	if exists == 0 {
		// UserCacheData의 초기 상태 정의
		initialData := defs.UserCacheData{
			ActiveClientSession: "",                                    // 빈 문자열 슬라이스로 초기화 (JSON: [])
			Account:             nil,                                   // 명시적으로 nil (JSON에서 생략 또는 null)
			AccountStats:        nil,                                   // 명시적으로 nil (JSON에서 생략 또는 null)
			SystemSaveData:      nil,                                   // 명시적으로 nil (JSON에서 생략 또는 null)
			SessionSaveData:     make(map[string]defs.SessionSaveData), // 빈 맵으로 초기화 (JSON: {})
		}

		// Redis에 저장
		err = SetJSON(ctx, redisKey, "$", initialData)
		if err != nil {
			logger.Error("캐시에서 키 '%s'를 찾을 수 없음", redisKey)
		}
		return err
	}

	return err
}

// db에서 가져온 유저 데이터 session:uuid / cachedata로 넣기
func StoreCacheData(ctx context.Context, uuid []byte, userData defs.UserCacheData) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	err := SetJSON(ctx, redisKey, "$", userData)

	if err != nil {
		return err
	}

	return nil
}

// uuid로 cachedata 제거
func DeleteCachedata(ctx context.Context, uuid []byte) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	err := Rdb.JSONDel(ctx, redisKey, "$").Err()

	if err != nil {
		return err
	}

	return nil

}
