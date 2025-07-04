package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pagefaultgames/rogueserver/defs"
)

// uuid로 한 유저의 cachedata가 있는지 확인
func IsValidCacheData(uuid []byte) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	exists, err := Rdb.Exists(Ctx, redisKey).Result()

	if exists == 0 {
		// UserCacheData의 초기 상태 정의
		initialData := defs.UserCacheData{
			ActiveClientSession: "",                                    // 빈 문자열 슬라이스로 초기화 (JSON: [])
			Account:             nil,                                   // 명시적으로 nil (JSON에서 생략 또는 null)
			AccountStats:        nil,                                   // 명시적으로 nil (JSON에서 생략 또는 null)
			SystemSaveData:      nil,                                   // 명시적으로 nil (JSON에서 생략 또는 null)
			SessionSaveData:     make(map[string]defs.SessionSaveData), // 빈 맵으로 초기화 (JSON: {})
		}

		// JSON으로 마샬링
		jsonDataBytes, err := json.Marshal(initialData)
		if err != nil {
			return fmt.Errorf("초기 UserCacheData JSON 마샬링 오류 (Key: %s): %s", redisKey, err)
		}
		jsonData := string(jsonDataBytes)

		// Redis에 저장
		err = Rdb.JSONSet(Ctx, redisKey, "$", jsonData).Err()
		log.Printf("캐시에서 키 '%s'를 찾을 수 없음", redisKey)
		return err
	}

	return err
}

// db에서 가져온 유저 데이터 session:uuid / cachedata로 넣기
func StoreCacheData(uuid []byte, userData defs.UserCacheData) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	err := Rdb.JSONSet(Ctx, redisKey, "$", userData).Err()

	if err != nil {
		return err
	}

	return nil
}

// uuid로 cachedata 제거
func DeleteCachedata(uuid []byte) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	err := Rdb.JSONDel(Ctx, redisKey, "$").Err()

	if err != nil {
		return err
	}

	return nil

}
