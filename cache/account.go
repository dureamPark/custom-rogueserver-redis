package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

// DB에서 가져온 AccountDBRow를 Redis 캐시에 저장하는 함수
func CacheAccountInRedis(dbRow defs.AccountDBRow) error {

	// Redis 키 생성: UUID (binary)를 16진수 문자열로 변환하고 접두사 추가
	redisKey := "session:" + base64.StdEncoding.EncodeToString(dbRow.UUID)

	// AccountDBRow를 AccountRedisData로 변환
	redisData := defs.AccountRedisData{
		Username:   dbRow.Username,
		Hash:       base64.StdEncoding.EncodeToString(dbRow.Hash),
		Salt:       base64.StdEncoding.EncodeToString(dbRow.Salt),
		Registered: dbRow.Registered,
		Banned:     dbRow.Banned,
	}

	if dbRow.LastLoggedIn.Valid {
		redisData.LastLoggedIn = &dbRow.LastLoggedIn.Time
	}
	if dbRow.LastActivity.Valid {
		redisData.LastActivity = &dbRow.LastActivity.Time
	}
	if dbRow.TrainerID.Valid {
		// smallint(5) unsigned는 0-65535 범위. uint16로 안전하게 변환 가능.
		val := uint16(dbRow.TrainerID.Int32)
		redisData.TrainerID = &val
	}
	if dbRow.SecretID.Valid {
		val := uint16(dbRow.SecretID.Int32)
		redisData.SecretID = &val
	}
	if dbRow.DiscordID.Valid {
		redisData.DiscordID = &dbRow.DiscordID.String
	}
	if dbRow.GoogleID.Valid {
		redisData.GoogleID = &dbRow.GoogleID.String
	}

	// JSON으로 마샬링
	jsonData, err := json.Marshal(redisData)
	if err != nil {
		log.Printf("Redis 데이터 JSON 마샬링 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// Redis에 저장
	err = SetJSON(Ctx, redisKey, "$.account", jsonData)
	if err != nil {
		log.Printf("Redis에 데이터 캐싱 오류 (키: %s): %s", redisKey, err)
		return err
	}

	log.Printf("계정 정보가 Redis에 성공적으로 캐시되었습니다. Key: %s, Expiration: %s\n", redisKey, sessionDataTTL)
	return nil
}

// CacheAccountStatsInRedis 함수는 AccountStatsData를 Redis에 캐시합니다.
// dbStats는 DB에서 읽어온 AccountStatsData 구조체입니다.
func CacheAccountStatsInRedis(dbStats defs.AccountStatsData) error {

	// Redis 키 생성
	redisKey := "session:" + base64.StdEncoding.EncodeToString(dbStats.UUID)

	// AccountStatsData를 AccountStatsRedisData로 변환 (UUID 제외)
	redisData := defs.AccountStatsRedisData{
		PlayTime:              dbStats.PlayTime,
		Battles:               dbStats.Battles,
		ClassicSessionsPlayed: dbStats.ClassicSessionsPlayed,
		SessionsWon:           dbStats.SessionsWon,
		HighestEndlessWave:    dbStats.HighestEndlessWave,
		HighestLevel:          dbStats.HighestLevel,
		PokemonSeen:           dbStats.PokemonSeen,
		PokemonDefeated:       dbStats.PokemonDefeated,
		PokemonCaught:         dbStats.PokemonCaught,
		PokemonHatched:        dbStats.PokemonHatched,
		EggsPulled:            dbStats.EggsPulled,
		RegularVouchers:       dbStats.RegularVouchers,
		PlusVouchers:          dbStats.PlusVouchers,
		PremiumVouchers:       dbStats.PremiumVouchers,
		GoldenVouchers:        dbStats.GoldenVouchers,
	}

	// Redis 저장용 데이터를 JSON으로 마샬링
	jsonData, err := json.Marshal(redisData)
	if err != nil {
		log.Printf("통계 데이터 JSON 마샬링 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// Redis에 JSON 데이터 저장
	err = SetJSON(Ctx, redisKey, "$.accountStats", jsonData)

	if err != nil {
		log.Printf("Redis에 통계 데이터 캐싱 오류 (키: %s): %s", redisKey, err)
		return err
	}

	log.Printf("계정 통계 데이터가 Redis에 성공적으로 캐시되었습니다. Key: %s, Expiration: %s\n", redisKey, sessionDataTTL)
	return nil
}

// session 활성화
func UpdateActiveSession(uuid []byte, sessionId string) error {

	log.Printf("UpdateActiveSession uuid : %s, sessionId : %s", base64.StdEncoding.EncodeToString(uuid), sessionId)
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	if sessionId == "" {
		return fmt.Errorf("sessionId is empty")
	}
	return SetJSON(Ctx, redisKey, "$.activeClientSession", fmt.Sprintf("\"%s\"", sessionId))
}

// 현재 session이 활성화되어 있는지 확인, 비활성화 시 새롭게 활성화
func IsActiveSession(uuid []byte, sessionId string) (bool, error) {
	//var id string
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	id, err := Rdb.JSONGet(Ctx, redisKey, ".activeClientSession").Result()

	if err != nil {
		// 초기화를 빈 문자열로 ""로 해서 확인하기
		err = UpdateActiveSession(uuid, sessionId)
		if err != nil {
			log.Printf("fail to Set Active Session in redis")
			return false, err
		}
		return true, nil
	}

	id = strings.Trim(id, "\"") // 쌍따옴표 제거
	log.Printf("id : %s, session id : %s", id, sessionId)
	return id == "" || id == sessionId, nil
}

// StoreSessionToken stores a token-uuid pair in Redis with TTL.
func StoreSessionToken(uuid []byte, token []byte) error {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	return Set(Ctx, key, uuid, sessionTokenTTL)
}

// FetchSessionToken retrieves the uuid for a given token from Redis.
func FetchSessionToken(token []byte) ([]byte, error) {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	return Rdb.Get(Ctx, key).Bytes()
}

// RemoveSessionFromToken removes the token-uuid mapping from Redis.
func RemoveSessionFromToken(token []byte) error {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	return Rdb.Del(Ctx, key).Err()
}

func FetchTrainerIds(uuid []byte) (int, int, error) {
	log.Println("FetchTrainerIds")
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// JSON.MGET을 사용하여 여러 경로의 값을 한 번에 가져올 수 있음
	// 결과는 []interface{} 형태의 슬라이스로 오며, 각 요소는 해당 경로의 값 또는 nil
	vals, err := Rdb.JSONMGet(Ctx, redisKey, "$.account.trainerId", "$.account.secretId").Result()

	if err == redis.Nil {
		return 0, 0, fmt.Errorf("캐시에서 UUID '%s'에 해당하는 계정을 찾을 수 없음: %w", uuid, err)
	} else if err != nil {
		return 0, 0, fmt.Errorf("RedisJSON.MGET 오류 (키: %s): %w", redisKey, err)
	}

	var trainerID int
	if vals[0] != nil {
		// RedisJSON은 숫자를 float64로 반환하는 경우가 많으므로 타입 변환 필요
		if tidFloat, ok := vals[0].(float64); ok {
			tidVal := int(tidFloat)
			trainerID = tidVal
		} else if tidInt, ok := vals[0].(int64); ok { // 또는 int64
			tidVal := int(tidInt)
			trainerID = tidVal
		}
	}

	var secretID int
	if vals[1] != nil {
		if sidFloat, ok := vals[1].(float64); ok {
			sidVal := int(sidFloat)
			secretID = sidVal
		} else if sidInt, ok := vals[1].(int64); ok {
			sidVal := int(sidInt)
			secretID = sidVal
		}
	}

	return trainerID, secretID, nil
}

// 트레이너 아이디 업데이트
func UpdateTrainerIds(trainerId, secretId int, uuid []byte) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// Redis 파이프라인을 사용하여 여러 명령을 원자적으로 (또는 더 효율적으로) 실행
	pipe := Rdb.Pipeline()

	// 2. trainerId 업데이트
	// JSON.SET key path value
	// path는 "$.trainerId"
	// value는 int 타입이므로 Redis가 JSON 숫자로 저장합니다.
	pipe.JSONSet(Ctx, redisKey, "$.account.trainerId", trainerId)

	// 3. secretId 업데이트
	pipe.JSONSet(Ctx, redisKey, "$.account.secretId", secretId)

	// 4. 파이프라인 실행
	cmders, err := pipe.Exec(Ctx)
	if err != nil {
		log.Printf("Redis 파이프라인 실행 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// 각 명령어의 성공 여부 확인 (선택적이지만 권장)
	for i, cmd := range cmders {
		if cmd.Err() != nil {
			fieldName := "trainerId"
			if i == 1 {
				fieldName = "secretId"
			}
			// 파이프라인 내의 특정 명령 실패 시 롤백 전략이 필요할 수 있으나,
			// 여기서는 일단 에러를 반환합니다.
			log.Printf("Redis 파이프라인 내 '%s' 업데이트 실패 (키: %s): %s", fieldName, redisKey, cmd.Err())
			return err
		}
		// log.Printf("Debug: Command %d result: %v", i, cmd.String())
	}

	log.Printf("키 %s의 trainerId가 %d로, secretId가 %d로 업데이트되었습니다.", redisKey, trainerId, secretId)
	return nil
}

func UpdateAccountLastActivity(uuid []byte) error {
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// 2. 현재 UTC 시간을 ISO 8601 형식 문자열로 준비
	// time.RFC3339Nano 또는 time.RFC3339 형식을 사용할 수 있습니다.
	// UTC_TIMESTAMP()와 가장 유사하게 하려면 Z (Zulu time)를 명시하는 것이 좋습니다.
	currentTimeStr := time.Now().UTC().Format(time.RFC3339Nano)
	// 또는 정밀도가 낮은 형식이 필요하면:
	// currentTimeStr := time.Now().UTC().Format(time.RFC3339) // 예: "2023-10-27T12:34:56Z"
	// 만약 Unix timestamp (숫자)로 저장한다면:
	// currentTimeUnix := time.Now().UTC().Unix()

	// 3. JSON.SET 명령으로 lastActivity 필드 업데이트
	// JSON.SET key path value
	// path는 "$.lastActivity"
	// value는 준비된 시간 문자열 (또는 숫자 타임스탬프)
	err := SetJSON(Ctx, redisKey, "$.account.lastActivity", currentTimeStr)
	if err != nil {
		log.Printf("Redis JSON.SET lastActivity 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// JSON.SET 결과 확인 (선택적)
	// log.Printf("JSON.SET 결과: %v", cmdResult)

	log.Printf("키 %s의 lastActivity가 '%s'로 업데이트되었습니다 (Redis).", redisKey, currentTimeStr)
	return nil
}

// UpdateAccountStatsInRedis 함수는 Redis에 저장된 계정 통계를 업데이트합니다.
func UpdateAccountStats(uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// Redis 파이프라인을 사용하여 여러 JSON.SET 명령을 효율적으로 실행
	pipe := Rdb.Pipeline()
	updateCount := 0

	// 2. `stats` (GameStats) 처리
	// DB 스키마의 컬럼 이름과 일치하는지 확인 (이전 코드의 columns 배열 사용)
	validStatColumns := []string{"playTime", "battles", "classicSessionsPlayed", "sessionsWon", "highestEndlessWave", "highestLevel", "pokemonSeen", "pokemonDefeated", "pokemonCaught", "pokemonHatched", "eggsPulled"}

	// GameStats가 map[string]interface{}라고 가정 (원래 코드 기반)
	m, ok := stats.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", stats)
	}

	for key, val := range m {
		if !slices.Contains(validStatColumns, key) {
			//log.Printf("경고: GameStats에 유효하지 않은 키 '%s'가 포함되어 무시합니다.", key)
			continue
		}

		// DB 스키마는 int(11)이었으므로, float64로 온 값을 int로 변환
		floatVal, ok := val.(float64)
		if !ok {
			log.Printf("경고: GameStats의 키 '%s'의 값이 float64가 아닙니다 (타입: %T). 무시합니다.", key, val)
			continue
		}
		intValue := int(floatVal) // 소수점 버림

		// JSONPath 생성 (예: "$.playTime")
		jsonPath := "$.accountStats." + key
		pipe.JSONSet(Ctx, redisKey, jsonPath, intValue)
		updateCount++
		// log.Printf("Debug: Pipelining JSON.SET %s %s %d", redisKey, jsonPath, intValue)
	}

	// 3. `voucherCounts` 처리
	voucherColumnMap := map[string]string{
		"0": "regularVouchers",
		"1": "plusVouchers",
		"2": "premiumVouchers",
		"3": "goldenVouchers",
	}
	for key, count := range voucherCounts {
		columnName, ok := voucherColumnMap[key]
		if !ok {
			log.Printf("경고: voucherCounts에 유효하지 않은 키 '%s'가 포함되어 무시합니다.", key)
			continue
		}
		jsonPath := "$.accountStats." + columnName
		pipe.JSONSet(Ctx, redisKey, jsonPath, count) // count는 이미 int
		updateCount++
		// log.Printf("Debug: Pipelining JSON.SET %s %s %d", redisKey, jsonPath, count)
	}

	// 4. 파이프라인 실행 (실제로 업데이트할 내용이 있을 때만)
	if updateCount == 0 {
		log.Printf("업데이트할 통계 또는 바우처 정보가 없습니다 (키: %s).", redisKey)
		return nil // 아무것도 안하고 성공
	}

	cmders, err := pipe.Exec(Ctx)
	if err != nil {
		log.Printf("Redis 파이프라인 실행 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// 각 명령어의 성공 여부 확인 (선택적)
	for i, cmd := range cmders {
		if cmd.Err() != nil {
			log.Printf("Redis 파이프라인 내 %d번째 업데이트 실패 (키: %s): %s", i+1, redisKey, cmd.Err())
			// 어떤 필드 업데이트가 실패했는지 특정하기 어려울 수 있음 (파이프라인 순서 기반 추정)
			return err
		}
	}

	log.Printf("키 %s의 계정 통계가 성공적으로 업데이트되었습니다 (업데이트된 필드 수: %d).", redisKey, updateCount)
	return nil
}
