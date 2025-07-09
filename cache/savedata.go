package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

// ReadSessionSaveData 함수는 Redis에서 특정 UUID와 슬롯에 해당하는 세션 저장 데이터를 읽어옵니다.
func ReadSessionSaveData(uuid []byte, slot int) (defs.SessionSaveData, error) { // defs.SessionSaveData로 변경해야 함
	var saveData defs.SessionSaveData // defs.SessionSaveData

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)
	redisKey := "session:" + encodedUUID

	jsonPath := fmt.Sprintf(`$.sessionSaveData["%s"]`, strconv.Itoa(slot))
	// Redis에서 JSON 데이터 가져오기
	jsonData, err := Rdb.JSONGet(Ctx, redisKey, jsonPath).Result()
	if err == redis.Nil {
		// 키가 존재하지 않는 경우, 빈 SessionSaveData와 함께 특정 에러 반환
		// 호출하는 쪽에서 이 에러를 식별하여 "새 게임" 또는 "슬롯 비어있음" 등으로 처리 가능
		log.Printf("세션 데이터를 찾을 수 없음 (키: %s): %s", redisKey, err)
		return saveData, err // err 대신 redis.Nil을 직접 전달하거나 커스텀 에러 사용
	} else if err != nil {
		// 그 외 Redis 오류
		log.Printf("Redis에서 세션 데이터 조회 오류 (키: %s): %s", redisKey, err)
		return saveData, err
	}

	// json 데이터 처리
	var saveDataArr defs.SessionSaveData
	if err := json.Unmarshal([]byte(jsonData), &saveDataArr); err != nil {
		log.Printf("세션 데이터 JSON 언마샬링 오류 (키: %s): %s", redisKey, err)
		return saveData, redis.Nil
	}

	return saveData, nil
}

// StoreSessionSaveData 함수는 주어진 SessionSaveData를 Redis에 저장합니다.
// 데이터는 JSON 형태로 저장되며, 만료 시간은 설정하지 않습니다 (필요시 추가 가능).
func StoreSessionSaveData(uuid []byte, data defs.SessionSaveData, slot int) error { // defs.SessionSaveData

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)
	redisKey := "session:" + encodedUUID

	jsonPath := fmt.Sprintf(`$.sessionSaveData["%s"]`, strconv.Itoa(slot))

	// SessionSaveData 구조체를 JSON으로 마샬링
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("세션 데이터 JSON 마샬링 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// Redis에 JSON 데이터 저장
	err = SetJSON(Ctx, redisKey, jsonPath, jsonData)
	if err != nil {
		log.Printf("Redis에 세션 데이터 저장 오류 (키: %s): %s", redisKey, err)
		return err
	}

	log.Printf("세션 데이터 저장 성공 (키: %s)", redisKey)
	return nil
}

// DeleteSessionSaveData 함수는 Redis에서 특정 UUID와 슬롯에 해당하는 세션 저장 데이터를 삭제합니다.
func DeleteSessionSaveData(uuid []byte, slot int) error {

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)
	redisKey := "session:" + encodedUUID

	jsonPath := fmt.Sprintf(`$.sessionSaveData["%s"]`, strconv.Itoa(slot))

	// Redis에서 해당 키 삭제
	result, err := Rdb.JSONDel(Ctx, redisKey, jsonPath).Result()
	if err != nil {
		return fmt.Errorf("Redis에서 세션 데이터 삭제 오류 (키: %s): %s", redisKey, err)
	}

	// 삭제 성공 여부 확인 (선택적이지만, 삭제 대상이 실제로 있었는지 확인 가능)
	if result == 0 {
		// 키가 존재하지 않아 아무것도 삭제되지 않은 경우.
		// 이를 에러로 처리할지, 아니면 성공으로 간주할지는 정책에 따라 다름.
		// 여기서는 일단 로그만 남기고 성공으로 처리.
		log.Printf("삭제할 세션 데이터가 Redis에 존재하지 않았습니다 (키: %s)", redisKey)
	} else {
		log.Printf("세션 데이터 삭제 성공 (키: %s, 삭제된 키 개수: %d)", redisKey, result)
	}

	return nil
}

func ReadSystemSaveData(uuid []byte)(defs.SystemSaveData, error){
	var systemData defs.SystemSaveData

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)
	redisKey := "session:" + encodedUUID

	jsonPath := fmt.Sprintf(`$.systemSaveData`)

	// Redis에서 JSON 데이터 가져오기
	jsonData, err := Rdb.JSONGet(Ctx, redisKey, jsonPath).Result()
	if err == redis.Nil {
		// 키가 존재하지 않는 경우, 빈 SessionSaveData와 함께 특정 에러 반환
		// 호출하는 쪽에서 이 에러를 식별하여 "새 게임" 또는 "슬롯 비어있음" 등으로 처리 가능
		log.Printf("시스템 데이터를 찾을 수 없음 (키: %s): %s", redisKey, err)
		return systemData, err // err 대신 redis.Nil을 직접 전달하거나 커스텀 에러 사용
	} else if err != nil {
		// 그 외 Redis 오류
		log.Printf("Redis에서 시스템 데이터 조회 오류 (키: %s): %s", redisKey, err)
		return systemData, err
	}

	if err := json.Unmarshal([]byte(jsonData), &systemData); err != nil {
		log.Printf("시스템 데이터 JSON 언마샬링 오류 (키: %s): %s", redisKey, err)
		//return systemData, err
	}

	return systemData, nil
}

// StoreSessionSaveData 함수는 주어진 SessionSaveData를 Redis에 저장합니다.
// 데이터는 JSON 형태로 저장되며, 만료 시간은 설정하지 않습니다 (필요시 추가 가능).
func StoreSystemSaveData(uuid []byte, data defs.SystemSaveData) error { // defs.SessionSaveData

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// SessionSaveData 구조체를 JSON으로 마샬링
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("세션 데이터 JSON 마샬링 오류 (키: %s): %s", redisKey, err)
	}

	// Redis에 JSON 데이터 저장
	err = SetJSON(Ctx, redisKey, "$.systemSaveData", jsonData)
	if err != nil {
		log.Printf("Redis에 세션 데이터 저장 오류 (키: %s): %s", redisKey, err)
		return err
	}

	log.Printf("세션 데이터 저장 성공 (키: %s)", redisKey)
	return nil
}

// FetchPlayTimeFromAccountStats 함수는 RedisJSON을 사용하여 캐시된 계정 통계에서 playTime만 가져옵니다.
// uuidBytes는 계정의 []byte UUID입니다.
func RetrievePlaytime(uuid []byte) (int, error) {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	rawResult, err := Rdb.JSONGet(Ctx, redisKey, ".accountStats.playTime").Result()
	if err != nil {
		log.Printf("Redis JSON.GET playTime 오류 (키: %s): %s", redisKey, err)
		return 0, err
	}

	var playTime int
	playTime, err = strconv.Atoi(rawResult)
	if err != nil {
		log.Printf("%s | playTime 값이 예상된 숫자 타입(float64 또는 int64)이 아님 (키: %s, 실제타입: %T, 값: %v)", err, redisKey, rawResult, rawResult)
		return 0, err
	}

	log.Printf("키 %s에서 PlayTime %d를 성공적으로 가져왔습니다.", redisKey, playTime)
	return playTime, nil
}
