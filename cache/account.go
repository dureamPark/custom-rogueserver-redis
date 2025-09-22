package cache

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/pagefaultgames/rogueserver/util/logger"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

func InitAccountStatsInRedis(ctx context.Context, uuid []byte) error {
	// Redis 키 생성
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// 기본 AccountStatsRedisData 생성
	redisData := defs.AccountStatsRedisData{
		PlayTime:              0,
		Battles:               0,
		ClassicSessionsPlayed: 0,
		SessionsWon:           0,
		HighestEndlessWave:    0,
		HighestLevel:          0,
		PokemonSeen:           0,
		PokemonDefeated:       0,
		PokemonCaught:         0,
		PokemonHatched:        0,
		EggsPulled:            0,
		RegularVouchers:       0,
		PlusVouchers:          0,
		PremiumVouchers:       0,
		GoldenVouchers:        0,
	}

	// Redis에 JSON 데이터 저장
	err := SetJSON(ctx, redisKey, "$.accountStats", redisData)

	if err != nil {
		logger.Error("Redis에 통계 데이터 캐싱 오류 (키: %s): %s", redisKey, err)
		return err
	}

	logger.Info("계정 통계 데이터가 Redis에 성공적으로 캐시되었습니다. Key: %s, Expiration: %s", redisKey, sessionDataTTL)
	return nil
}

// DB에서 가져온 AccountDBRow를 Redis 캐시에 저장하는 함수
func CacheAccountInRedis(ctx context.Context, dbRow defs.AccountDBRow) error {

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

	// Redis에 저장
	err := SetJSON(ctx, redisKey, "$.account", redisData)
	if err != nil {
		logger.Error("Redis에 데이터 캐싱 오류 (키: %s): %s", redisKey, err)
		return err
	}

	logger.Info("계정 정보가 Redis에 성공적으로 캐시되었습니다. Key: %s, Expiration: %s\n", redisKey, sessionDataTTL)
	return nil
}

// CacheAccountStatsInRedis 함수는 AccountStatsData를 Redis에 캐시합니다.
// dbStats는 DB에서 읽어온 AccountStatsData 구조체입니다.
func CacheAccountStatsInRedis(ctx context.Context, uuid []byte, dbStats defs.AccountStatsData) error {

	// Redis 키 생성
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

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

	// Redis에 JSON 데이터 저장
	err := SetJSON(ctx, redisKey, "$.accountStats", redisData)

	if err != nil {
		logger.Error("Redis에 통계 데이터 캐싱 오류 (키: %s): %s", redisKey, err)
		return err
	}

	logger.Info("계정 통계 데이터가 Redis에 성공적으로 캐시되었습니다. Key: %s, Expiration: %s", redisKey, sessionDataTTL)
	return nil
}

func FetchUsernameBySessionToken(ctx context.Context, token []byte) (string, error) {
	key := "token:" + string(token)
	// uuid 찾기
	uuid, err := Rdb.Get(ctx, key).Result()

	if err != nil {
		return "", err
	}

	// username 찾기
	sessionkey := "session:" + uuid
	username, err := Rdb.JSONGet(ctx, sessionkey, "$.account.username").Result()

	return username, err
}

// session 활성화
func UpdateActiveSession(ctx context.Context, uuid []byte, sessionId string) error {

	logger.Info("UpdateActiveSession uuid : %s, sessionId : %s", base64.StdEncoding.EncodeToString(uuid), sessionId)
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	if sessionId == "" {
		return fmt.Errorf("sessionId is empty")
	}
	// expected value at line 1 column 1  문자열 형식이 안맞아서 ""를 붙여주기
	return SetJSON(ctx, redisKey, "$.activeClientSession", fmt.Sprintf("\"%s\"", sessionId))
}

// 현재 session이 활성화되어 있는지 확인, 비활성화 시 새롭게 활성화
func IsActiveSession(ctx context.Context, uuid []byte, sessionId string) (bool, error) {
	//var id
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	id, err := Rdb.JSONGet(ctx, redisKey, ".activeClientSession").Result()

	if err != nil {
		logger.Error("fail to Set Active Session in redis")
		// 초기화를 빈 문자열로 ""로 해서 확인하기
		err = UpdateActiveSession(ctx, uuid, sessionId)
		if err != nil {
			logger.Error("fail to Set Active Session in redis : %s, session id : %s, get id : %s", err, sessionId, id)
			return false, err
		}
		return true, nil
	}

	var cur string
	if err := json.Unmarshal([]byte(id), &cur); err != nil {

		// 비어있는 경우 = 첫 로그인인 경우는 패쓰
		if id != "" {
			return false, fmt.Errorf("unmarshal error (%s): %w", id, err)
		}
	}

	// .activeClientSession로 가져오면 문자열의 경우 ""도 같이 가져옴
	// id = strings.Trim(id, "") // 쌍따옴표 제거
	// id = strings.Trim(id, "") // 두 번..?
	// // if(!(id == "" || id == sessionId)){
	// // 	logger.Error("id : %s, session id : %s", id, sessionId)

	// // }

	logger.Info("session id : %s, get id : %s", sessionId, cur)

	return cur == "" || cur == sessionId, nil
}

// StoreSessionToken stores a token-uuid pair in Redis with TTL.
func StoreSessionToken(ctx context.Context, uuid []byte, token []byte) error {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	return Set(ctx, key, uuid, sessionTokenTTL)
}

// FetchUUIDFromToken retrieves the uuid for a given token from Redis.
func FetchUUIDFromToken(ctx context.Context, token []byte) ([]byte, error) {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	return Rdb.Get(ctx, key).Bytes()
}

func FetchUsernameFromUUID(ctx context.Context, uuid []byte) (string, error) {

	key := "session:" + base64.StdEncoding.EncodeToString(uuid)
	username, err := Rdb.Get(ctx, key).Result()

	return username, err
}

// RemoveSessionFromToken removes the token-uuid mapping from Redis.
func RemoveSessionFromToken(ctx context.Context, token []byte) error {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	return Rdb.Del(ctx, key).Err()
}

func FetchTrainerIds(ctx context.Context, uuid []byte) (int, int, error) {
	logger.Info("FetchTrainerIds")
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// JSON.MGET을 사용하여 여러 경로의 값을 한 번에 가져올 수 있음
	// 결과는 []interface{} 형태의 슬라이스로 오며, 각 요소는 해당 경로의 값 또는 nil
	result, err := Rdb.JSONGet(ctx, redisKey, "$.account.trainerId", "$.account.secretId").Result()

	if err == redis.Nil {
		return 0, 0, fmt.Errorf("캐시에서 UUID '%s'에 해당하는 계정을 찾을 수 없음: %w", uuid, err)
	} else if err != nil {
		return 0, 0, fmt.Errorf("RedisJSON.MGET 오류 (키: %s): %w", redisKey, err)
	}

	var values []int
	err = json.Unmarshal([]byte(result), &values)
	if err != nil {
		// 에러 처리
	}

	trainerID := 0
	secretID := 0

	if len(values) >= 2 {
		trainerID = values[0]
		secretID = values[1]
	}

	return trainerID, secretID, nil
}

// 트레이너 아이디 업데이트
func UpdateTrainerIds(ctx context.Context, trainerId, secretId int, uuid []byte) error {

	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)

	// Redis 파이프라인을 사용하여 여러 명령을 원자적으로 (또는 더 효율적으로) 실행
	pipe := Rdb.Pipeline()

	// 2. trainerId 업데이트
	// JSON.SET key path value
	// path는 "$.trainerId"
	// value는 int 타입이므로 Redis가 JSON 숫자로 저장합니다.

	// 객체가 없을 때는 생성해서 넣어주기
	pipe.Do(ctx, "JSON.SET", redisKey, "$", "{}", "NX")
	pipe.Do(ctx, "JSON.SET", redisKey, "$.account", "{}", "NX")

	pipe.JSONSet(ctx, redisKey, "$.account.trainerId", trainerId)

	// 3. secretId 업데이트
	pipe.JSONSet(ctx, redisKey, "$.account.secretId", secretId)

	// 4. 파이프라인 실행
	cmders, err := pipe.Exec(ctx)
	// if err != nil {
	// 	logger.Error("Redis 파이프라인 실행 오류 (키: %s): %s", redisKey, err)
	// 	return err
	// }

	// 각 명령어의 성공 여부 확인 (선택적이지만 권장)
	for i, cmd := range cmders {
		if cmd.Err() != nil && cmd.Err() != redis.Nil {
			logger.Error("명령 %d 실패: %v", i, cmd.Err())
			return err
		}
	}
	logger.Info("키 %s의 trainerId가 %d로, secretId가 %d로 업데이트되었습니다.", redisKey, trainerId, secretId)
	return nil
}

func UpdateAccountLastActivity(ctx context.Context, uuid []byte) error {
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
	Rdb.Do(ctx, "JSON.SET", redisKey, "$.account", "{}", "NX")
	err := SetJSON(ctx, redisKey, "$.account.lastActivity", currentTimeStr)
	if err != nil {
		logger.Error("Redis JSON.SET lastActivity 오류 (키: %s): %s", redisKey, err)
		return err
	}

	// JSON.SET 결과 확인 (선택적)
	// log.Printf("JSON.SET 결과: %v", cmdResult)

	logger.Info("키 %s의 lastActivity가 '%s'로 업데이트되었습니다 (Redis).", redisKey, currentTimeStr)
	return nil
}

// UpdateAccountStatsInRedis 함수는 Redis에 저장된 계정 통계를 업데이트합니다.
func UpdateAccountStats(ctx context.Context, uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {

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

	pipe.Do(ctx, "JSON.SET", redisKey, "$", "{}", "NX")
	pipe.Do(ctx, "JSON.SET", redisKey, "$.accountStats", "{}", "NX")
	for key, val := range m {
		if !slices.Contains(validStatColumns, key) {
			//logger.Warn("경고: GameStats에 유효하지 않은 키 '%s'가 포함되어 무시합니다.", key)
			continue
		}

		// DB 스키마는 int(11)이었으므로, float64로 온 값을 int로 변환
		floatVal, ok := val.(float64)
		if !ok {
			logger.Warn("경고: GameStats의 키 '%s'의 값이 float64가 아닙니다 (타입: %T). 무시합니다.", key, val)
			continue
		}
		intValue := int(floatVal) // 소수점 버림

		// JSONPath 생성 (예: "$.playTime")
		jsonPath := "$.accountStats." + key
		pipe.Do(ctx, "JSON.SET", redisKey, jsonPath, "{}", "NX")
		pipe.JSONSet(ctx, redisKey, jsonPath, intValue)
		updateCount++
		logger.Info("Debug: Pipelining JSON.SET %s %s %d", redisKey, jsonPath, intValue)
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
			logger.Warn("경고: voucherCounts에 유효하지 않은 키 '%s'가 포함되어 무시합니다.", key)
			continue
		}
		jsonPath := "$.accountStats." + columnName
		pipe.Do(ctx, "JSON.SET", redisKey, jsonPath, "{}", "NX")
		pipe.JSONSet(ctx, redisKey, jsonPath, count) // count는 이미 int
		updateCount++
		logger.Info("Debug: Pipelining JSON.SET %s %s %d", redisKey, jsonPath, count)
	}

	// 4. 파이프라인 실행 (실제로 업데이트할 내용이 있을 때만)
	if updateCount == 0 {
		logger.Info("업데이트할 통계 또는 바우처 정보가 없습니다 (키: %s).", redisKey)
		return nil // 아무것도 안하고 성공
	}

	cmders, err := pipe.Exec(ctx)
	// if err != nil {
	// 	logger.Error("Redis 파이프라인 실행 오류 (키: %s): %s", redisKey, err)
	// 	return err
	// }

	// 각 명령어의 성공 여부 확인 (선택적)
	for i, cmd := range cmders {
		if cmd.Err() != nil && cmd.Err() != redis.Nil {
			logger.Error("Redis 파이프라인 내 %d번째 업데이트 실패 (cmd: %s): %s", i+1, cmd, cmd.Err())
			// 어떤 필드 업데이트가 실패했는지 특정하기 어려울 수 있음 (파이프라인 순서 기반 추정)
			return err
		}
	}

	logger.Info("키 %s의 계정 통계가 성공적으로 업데이트되었습니다 (업데이트된 필드 수: %d).", redisKey, updateCount)
	return nil
}
