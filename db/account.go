/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package db

import (
	"database/sql"
	"encoding/base64"

	//"encoding/base64"
	"errors"
	"fmt"

	//"log"
	"slices"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pagefaultgames/rogueserver/defs"
	//"github.com/pagefaultgames/rogueserver/metrics"
	//redis "github.com/redis/go-redis/v9"
)

// DB에서 uuid로 accounts 정보를 모두 가져오는 함수
func GetAccountFromDB(uuid []byte) (defs.AccountDBRow, error) {
	var account defs.AccountDBRow

	query := `
		SELECT uuid, username, hash, salt, registered, lastLoggedIn, 
		       lastActivity, banned, trainerId, secretId, discordId, googleId 
		FROM accounts 
		WHERE uuid = ?
	`
	// MariaDB/MySQL 드라이버는 '?'를 파라미터 플레이스홀더로 사용
	// uuidBytes는 []byte 타입이므로 DB의 binary(16)과 직접 비교 가능

	row := handle.QueryRow(query, uuid)
	err := row.Scan(
		&account.UUID,
		&account.Username,
		&account.Hash,
		&account.Salt,
		&account.Registered,
		&account.LastLoggedIn,
		&account.LastActivity,
		&account.Banned,
		&account.TrainerID,
		&account.SecretID,
		&account.DiscordID,
		&account.GoogleID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 결과가 없는 경우는 에러로 처리 (호출하는 쪽에서 이 에러를 구분하여 처리 가능)
			return account, fmt.Errorf("UUID '%s'에 해당하는 계정을 찾을 수 없음: %w", uuid, err)
		}
		// 그 외 DB 스캔 오류 등
		return account, fmt.Errorf("DB에서 계정 조회 중 오류 (UUID: %s): %w", uuid, err)
	}

	return account, nil
}

// GetAccountStatsFromDB 함수는 DB에서 특정 UUID에 해당하는 accountStats 데이터를 가져옵니다.
// uuidBytes는 []byte 타입의 UUID입니다.
func GetAccountStatsFromDB(uuidBytes []byte) (defs.AccountStatsData, error) { // defs.AccountStatsData
	var stats defs.AccountStatsData // defs.AccountStatsData

	// uuidBytes가 nil이거나 길이가 맞는지 기본 검사 (선택적)
	if uuidBytes == nil || len(uuidBytes) != 16 {
		return stats, fmt.Errorf("잘못된 UUID 바이트 배열입니다. nil이거나 길이가 16이 아닙니다.")
	}

	query := `
		SELECT 
			uuid, playTime, battles, classicSessionsPlayed, sessionsWon, 
			highestEndlessWave, highestLevel, pokemonSeen, pokemonDefeated, 
			pokemonCaught, pokemonHatched, eggsPulled, regularVouchers, 
			plusVouchers, premiumVouchers, goldenVouchers 
		FROM accountStats 
		WHERE uuid = ?
	`
	// DB 드라이버는 '?'를 파라미터 플레이스홀더로 사용
	// uuidBytes는 []byte 타입이므로 DB의 binary(16)과 직접 비교 가능

	row := handle.QueryRow(query, uuidBytes)
	err := row.Scan(
		&stats.UUID, // []byte로 스캔
		&stats.PlayTime,
		&stats.Battles,
		&stats.ClassicSessionsPlayed,
		&stats.SessionsWon,
		&stats.HighestEndlessWave,
		&stats.HighestLevel,
		&stats.PokemonSeen,
		&stats.PokemonDefeated,
		&stats.PokemonCaught,
		&stats.PokemonHatched,
		&stats.EggsPulled,
		&stats.RegularVouchers,
		&stats.PlusVouchers,
		&stats.PremiumVouchers,
		&stats.GoldenVouchers,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 결과가 없는 경우. accountStats 테이블은 기본값이 0인 NOT NULL 필드로 구성되어 있으므로,
			// 레코드가 없다면 '새 사용자' 또는 '통계 없음'으로 간주하고
			// 모든 필드가 0인 AccountStatsData (제로 값 구조체)를 반환할 수 있습니다.
			// 또는 특정 에러를 반환하여 호출자가 구분하도록 할 수 있습니다.
			// 여기서는 에러를 반환합니다.
			return stats, fmt.Errorf("UUID '%s'에 해당하는 통계 데이터를 찾을 수 없음: %w", base64.StdEncoding.EncodeToString(uuidBytes), err)
		}
		// 그 외 DB 스캔 오류 등
		return stats, fmt.Errorf("DB에서 통계 데이터 조회 중 오류 (UUID: %s): %w", base64.StdEncoding.EncodeToString(uuidBytes), err)
	}

	return stats, nil
}

func AddAccountRecord(uuid []byte, username string, key, salt []byte) error {
	_, err := handle.Exec("INSERT INTO accounts (uuid, username, hash, salt, registered) VALUES (?, ?, ?, ?, UTC_TIMESTAMP())", uuid, username, key, salt)
	if err != nil {
		return err
	}

	return nil
}

/*
func AddAccountSession(username string, token []byte) error {
    ctx := cache.Ctx

    // 1) Redis에 session:token{username} 저장
    tokenStr := base64.StdEncoding.EncodeToString(token)
    userKey := fmt.Sprintf("session:token:%s", username)
    if err := cache.Rdb.Set(ctx, userKey, tokenStr, sessionTTL).Err(); err != nil {
        return fmt.Errorf("redis SET token error: %w", err)
    }

    // 2) Redis에 session:uuid{token} 저장 (인증용 맵핑)
    var uuid []byte
    if err := handle.QueryRow("SELECT uuid FROM accounts WHERE username = ?", username).Scan(&uuid); err != nil {
        return fmt.Errorf("fetch uuid: %w", err)
    }
    uuidKey := fmt.Sprintf("session:uuid:%s", tokenStr)
    if err := cache.Rdb.Set(ctx, uuidKey, string(uuid), sessionTTL).Err(); err != nil {
        return fmt.Errorf("redis SET uuid error: %w", err)
    }

    // 3) Dirty Set에 username 추가 → 나중에 워커가 꺼내서 DB에 반영
    if err := cache.Rdb.SAdd(ctx, "dirty:sessions", username).Err(); err != nil {
        return fmt.Errorf("redis SAdd dirty error: %w", err)
    }

    return nil
}
*/

// func AddAccountSession(username string, token []byte) error { //로그인 시 username에 token 저장해주는 함수.
// 	ctx := cache.Ctx

// 	// 캐시 조회 과정
// 	cacheKey := fmt.Sprintf(sessionTokenKeyFmt, username) //key:value=username:token
// 	s, err := cache.Rdb.Get(ctx, cacheKey).Result()       //cache hit or cache miss

// 	log.Printf("cache getttt")

// 	if err == nil {
// 		//cache hit인 경우.
// 		metrics.CacheHits.Inc() //prometheus에서 cache miss 확인하기 위한 준비 중..
// 		log.Printf("[CACHE HIT]   key=%s", cacheKey)
// 		//redis 에 저장할 때 문제가 발생하지 않도록 인코딩 및 디코딩을 진행하여 사용.
// 		if decoded, decErr := base64.StdEncoding.DecodeString(s); decErr == nil && len(decoded) == len(token) {
// 			log.Printf("token copy")
// 			copy(token, decoded)
// 			return nil
// 		}
// 		// 디코드 실패 시 DB 로직으로 넘어감
// 	} else if err != redis.Nil {
// 		// ── CACHE ERROR ──
// 		log.Printf("redis error")
// 		return fmt.Errorf("redis GET error: %w", err)
// 	}

// 	// err == redis.Nil → CACHE MISS인 경우.
// 	log.Printf("[CACHE MISS]  key=%s", cacheKey)

// 	metrics.CacheMisses.Inc() //cache miss 확인하기 위한 준비 중..

// 	// DB data 추가 sessions 테이블->해당 과정은 write-back 구조와 맞지 않음.
// 	//주석 처리해두고 추후에 write-back에 맞게 구조 수정할 예정.
// 	/*
// 	   log.Printf("DB insert before")
// 	   if _, err := handle.Exec(`INSERT INTO sessions (uuid, token, expire) SELECT a.uuid, ?, DATE_ADD(UTC_TIMESTAMP(), INTERVAL 1 WEEK) FROM accounts a WHERE a.username = ?`, token, username,); err != nil {
// 	       log.Printf("DB insert error: %v", err)
// 	       return err
// 	   }

// 	   // 3) DB 업데이트: lastLoggedIn **업데이트부분은 일단 pass**
// 	   log.Printf("DB update before")
// 	   if _, err := handle.Exec(`UPDATE accounts SET lastLoggedIn = UTC_TIMESTAMP() WHERE username = ?`, username,); err != nil {
// 	       log.Printf("DB update error: %v", err)
// 	       return err
// 	   }
// 	*/

// 	//캐시에 저장하는 부분 구현.
// 	log.Printf("cache save before")
// 	tokenStr := base64.StdEncoding.EncodeToString(token)

// 	//1.username → token, username과 token, TTL을 cache로 설정.
// 	if err := cache.Rdb.Set(ctx, cacheKey, tokenStr, sessionTTL).Err(); err != nil {
// 		return fmt.Errorf("redis SET token error: %w", err)
// 	}
// 	log.Printf("username -> token")

// 	metrics.CacheHits.Inc()

// 	//2.token → uuid 역매핑
// 	// DB에서 uuid를 다시 조회해서 저장 -> uuid설정이 없으면 로그인 이후 페이지로 이동이
// 	//안 되는 문제가 있음. 해당 문제 원인 파악 중..
// 	var uuid []byte
// 	if err := handle.QueryRow(
// 		"SELECT uuid FROM accounts WHERE username = ?", username).Scan(&uuid); err != nil {
// 		return fmt.Errorf("fetch uuid for cache: %w", err)
// 	}
// 	uuidKey := fmt.Sprintf(sessionUUIDKeyFmt, tokenStr)
// 	if err := cache.Rdb.Set(ctx, uuidKey, string(uuid), sessionTTL).Err(); err != nil {
// 		return fmt.Errorf("redis SET uuid error: %w", err)
// 	}

// 	log.Printf("return nil")
// 	return nil
// }

//아래가 원본 함수.

func AddAccountSession(username string, token []byte) error {
	_, err := handle.Exec("INSERT INTO sessions (uuid, token, expire) SELECT a.uuid, ?, DATE_ADD(UTC_TIMESTAMP(), INTERVAL 1 WEEK) FROM accounts a WHERE a.username = ?", token, username)
	if err != nil {
		return err
	}

	_, err = handle.Exec("UPDATE accounts SET lastLoggedIn = UTC_TIMESTAMP() WHERE username = ?", username)
	if err != nil {
		return err
	}

	return nil
}

func AddDiscordIdByUsername(discordId string, username string) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = ? WHERE username = ?", discordId, username)
	if err != nil {
		return err
	}

	return nil
}

func AddGoogleIdByUsername(googleId string, username string) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = ? WHERE username = ?", googleId, username)
	if err != nil {
		return err
	}

	return nil
}

func AddGoogleIdByUUID(googleId string, uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = ? WHERE uuid = ?", googleId, uuid)
	if err != nil {
		return err
	}

	return nil
}

func AddDiscordIdByUUID(discordId string, uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = ? WHERE uuid = ?", discordId, uuid)
	if err != nil {
		return err
	}

	return nil
}

func FetchUsernameByDiscordId(discordId string) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE discordId = ?", discordId).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func FetchUsernameByGoogleId(googleId string) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE googleId = ?", googleId).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func FetchDiscordIdByUsername(username string) (string, error) {
	var discordId sql.NullString
	err := handle.QueryRow("SELECT discordId FROM accounts WHERE username = ?", username).Scan(&discordId)
	if err != nil {
		return "", err
	}

	if !discordId.Valid {
		return "", nil
	}

	return discordId.String, nil
}

func FetchGoogleIdByUsername(username string) (string, error) {
	var googleId sql.NullString
	err := handle.QueryRow("SELECT googleId FROM accounts WHERE username = ?", username).Scan(&googleId)
	if err != nil {
		return "", err
	}

	if !googleId.Valid {
		return "", nil
	}

	return googleId.String, nil
}

func FetchDiscordIdByUUID(uuid []byte) (string, error) {
	var discordId sql.NullString
	err := handle.QueryRow("SELECT discordId FROM accounts WHERE uuid = ?", uuid).Scan(&discordId)
	if err != nil {
		return "", err
	}

	if !discordId.Valid {
		return "", nil
	}

	return discordId.String, nil
}

func FetchGoogleIdByUUID(uuid []byte) (string, error) {
	var googleId sql.NullString
	err := handle.QueryRow("SELECT googleId FROM accounts WHERE uuid = ?", uuid).Scan(&googleId)
	if err != nil {
		return "", err
	}

	if !googleId.Valid {
		return "", nil
	}

	return googleId.String, nil
}

func FetchUsernameBySessionToken(token []byte) (string, error) {
	var username string
	err := handle.QueryRow("SELECT a.username FROM accounts a JOIN sessions s ON a.uuid = s.uuid WHERE s.token = ?", token).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func CheckUsernameExists(username string) (string, error) {
	var dbUsername sql.NullString
	err := handle.QueryRow("SELECT username FROM accounts WHERE username = ?", username).Scan(&dbUsername)
	if err != nil {
		return "", err
	}
	if !dbUsername.Valid {
		return "", nil
	}

	return dbUsername.String, nil
}

func FetchLastLoggedInDateByUsername(username string) (string, error) {
	var lastLoggedIn sql.NullString
	err := handle.QueryRow("SELECT lastLoggedIn FROM accounts WHERE username = ?", username).Scan(&lastLoggedIn)
	if err != nil {
		return "", err
	}
	if !lastLoggedIn.Valid {
		return "", nil
	}

	return lastLoggedIn.String, nil
}

type AdminSearchResponse struct {
	Username     string `json:"username"`
	DiscordId    string `json:"discordId"`
	GoogleId     string `json:"googleId"`
	LastActivity string `json:"lastLoggedIn"` // TODO: this is currently lastLoggedIn to match server PR #54 with pokerogue PR #4198. We're hotfixing the server with this PR to return lastActivity, but we're not hotfixing the client, so are leaving this as lastLoggedIn so that it still talks to the client properly
	Registered   string `json:"registered"`
}

func FetchAdminDetailsByUsername(dbUsername string) (AdminSearchResponse, error) {
	var username, discordId, googleId, lastActivity, registered sql.NullString
	var adminResponse AdminSearchResponse

	err := handle.QueryRow("SELECT username, discordId, googleId, lastActivity, registered from accounts WHERE username = ?", dbUsername).Scan(&username, &discordId, &googleId, &lastActivity, &registered)
	if err != nil {
		return adminResponse, err
	}

	adminResponse = AdminSearchResponse{
		Username:     username.String,
		DiscordId:    discordId.String,
		GoogleId:     googleId.String,
		LastActivity: lastActivity.String,
		Registered:   registered.String,
	}

	return adminResponse, nil
}

func UpdateAccountPassword(uuid, key, salt []byte) error {
	_, err := handle.Exec("UPDATE accounts SET (hash, salt) VALUES (?, ?) WHERE uuid = ?", key, salt, uuid)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAccountLastActivity(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET lastActivity = UTC_TIMESTAMP() WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAccountStats(uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {
	var columns = []string{"playTime", "battles", "classicSessionsPlayed", "sessionsWon", "highestEndlessWave", "highestLevel", "pokemonSeen", "pokemonDefeated", "pokemonCaught", "pokemonHatched", "eggsPulled", "regularVouchers", "plusVouchers", "premiumVouchers", "goldenVouchers"}

	var statCols []string
	var statValues []interface{}

	m, ok := stats.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", stats)
	}

	for k, v := range m {
		value, ok := v.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", v)
		}

		if slices.Contains(columns, k) {
			statCols = append(statCols, k)
			statValues = append(statValues, value)
		}
	}

	for k, v := range voucherCounts {
		var column string
		switch k {
		case "0":
			column = "regularVouchers"
		case "1":
			column = "plusVouchers"
		case "2":
			column = "premiumVouchers"
		case "3":
			column = "goldenVouchers"
		default:
			continue
		}
		statCols = append(statCols, column)
		statValues = append(statValues, v)
	}

	var statArgs []interface{}
	statArgs = append(statArgs, uuid)
	for range 2 {
		statArgs = append(statArgs, statValues...)
	}

	query := "INSERT INTO accountStats (uuid"

	for _, col := range statCols {
		query += ", " + col
	}

	query += ") VALUES (?"

	for range len(statCols) {
		query += ", ?"
	}

	query += ") ON DUPLICATE KEY UPDATE "

	for i, col := range statCols {
		if i > 0 {
			query += ", "
		}

		query += col + " = ?"
	}

	_, err := handle.Exec(query, statArgs...)
	if err != nil {
		return err
	}

	return nil
}

func SetAccountBanned(uuid []byte, banned bool) error {
	_, err := handle.Exec("UPDATE accounts SET banned = ? WHERE uuid = ?", banned, uuid)
	if err != nil {
		return err
	}

	return nil
}

func FetchAccountKeySaltFromUsername(username string) ([]byte, []byte, error) {
	var key, salt []byte
	err := handle.QueryRow("SELECT hash, salt FROM accounts WHERE username = ?", username).Scan(&key, &salt)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func FetchTrainerIds(uuid []byte) (trainerId, secretId int, err error) {
	err = handle.QueryRow("SELECT trainerId, secretId FROM accounts WHERE uuid = ?", uuid).Scan(&trainerId, &secretId)
	if err != nil {
		return 0, 0, err
	}

	return trainerId, secretId, nil
}

func UpdateTrainerIds(trainerId, secretId int, uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET trainerId = ?, secretId = ? WHERE uuid = ?", trainerId, secretId, uuid)
	if err != nil {
		return err
	}

	return nil
}

func IsActiveSession(uuid []byte, sessionId string) (bool, error) {
	var id string
	err := handle.QueryRow("SELECT clientSessionId FROM activeClientSessions WHERE uuid = ?", uuid).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = UpdateActiveSession(uuid, sessionId)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		return false, err
	}

	return id == "" || id == sessionId, nil
}

func UpdateActiveSession(uuid []byte, clientSessionId string) error {
	_, err := handle.Exec("INSERT INTO activeClientSessions (uuid, clientSessionId) VALUES (?, ?) ON DUPLICATE KEY UPDATE clientSessionId = ?", uuid, clientSessionId, clientSessionId)
	if err != nil {
		return err
	}

	return nil
}

// func FetchUUIDFromToken(token []byte) ([]byte, error) { //uuid를 미리 cache에 설정하지 않으면
// 	//문제가 발생하는 것으로 추정되던 곳
// 	//token을 가지고 있는 uuid가 누구인지 찾는 함수.
// 	ctx := cache.Ctx
// 	tokenStr := base64.StdEncoding.EncodeToString(token)
// 	cacheKey := fmt.Sprintf(sessionUUIDKeyFmt, tokenStr)

// 	log.Printf("cache before")

// 	// 1.cache 조회
// 	if u, err := cache.Rdb.Get(ctx, cacheKey).Result(); err == nil {
// 		// cache hit
// 		log.Printf("cache hit")
// 		return []byte(u), nil
// 	} else if err != nil && err != redis.Nil {
// 		// 실제 Redis 에러
// 		log.Printf("redis error")
// 		return nil, fmt.Errorf("redis GET error: %w", err)
// 	}

// 	// err == redis.Nil 일 때만 아래로 (cache miss)
// 	log.Printf("cache miss")

// 	// 2.DB 조회
// 	var uuid []byte
// 	err := handle.QueryRow("SELECT uuid FROM sessions WHERE token = ?", token).Scan(&uuid)

// 	log.Printf("query complete")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 3.Redis에 캐시 설정. 다음 인증부터는 무조건적으로 hit 발생하도록.
// 	cache.Rdb.Set(ctx, cacheKey, string(uuid), sessionTTL)
// 	return uuid, nil
// }

// 위 함수의 기존 함수.
func FetchUUIDFromToken(token []byte) ([]byte, error) {
	var uuid []byte
	//user info DB에서 조회. 따라서 cache setting 필요.
	err := handle.QueryRow("SELECT uuid FROM sessions WHERE token = ?", token).Scan(&uuid)
	if err != nil {
		return nil, err
	}

	return uuid, nil
}

func RemoveSessionFromToken(token []byte) error {
	_, err := handle.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		return err
	}

	return nil
}

func FetchUsernameFromUUID(uuid []byte) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE uuid = ?", uuid).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func FetchUUIDFromUsername(username string) ([]byte, error) {
	var uuid []byte
	err := handle.QueryRow("SELECT uuid FROM accounts WHERE username = ?", username).Scan(&uuid)
	if err != nil {
		return nil, err
	}

	return uuid, nil
}

func RemoveDiscordIdByUUID(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = NULL WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func RemoveGoogleIdByUUID(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = NULL WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func RemoveGoogleIdByUsername(username string) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}

	return nil
}

func RemoveDiscordIdByUsername(username string) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}

	return nil
}

func RemoveDiscordIdByDiscordId(discordId string) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = NULL WHERE discordId = ?", discordId)
	if err != nil {
		return err
	}

	return nil
}

func RemoveGoogleIdByDiscordId(discordId string) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = NULL WHERE discordId = ?", discordId)
	if err != nil {
		return err
	}

	return nil
}
