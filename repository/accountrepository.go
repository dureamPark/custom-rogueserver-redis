package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"
	"github.com/redis/go-redis/v9"
)

const sessionDataTTL = time.Hour * 24 * 7
const sessionTokenTTL = time.Hour * 24 * 7

// accountCacheRepository는 AccountRepository 인터페이스의 캐시 구현체입니다.
type accountCacheRepository struct {
	cache *redis.Client
	next  AccountRepository // 다음 계층 (DB 레포지토리)
}

// NewAccountCacheRepository는 캐시와 통신하는 새로운 account repository를 생성합니다.
func NewAccountCacheRepository(redisClient *redis.Client, nextRepo AccountRepository) AccountRepository {
	return &accountCacheRepository{cache: redisClient, next: nextRepo}
}

func (r *accountCacheRepository) GetAccount(ctx context.Context, uuid []byte) (defs.AccountDBRow, error) {
	return r.next.GetAccount(ctx, uuid)
}

func (r *accountCacheRepository) GetAccountStats(ctx context.Context, uuid []byte) (defs.AccountStatsData, error) {
	// db에서 로그인 유저 통계 정보 가져와서 cache에 저장
	accountStatsData, err := r.next.GetAccountStats(ctx, uuid)
	logger.Info("Login account StatsData : %s", accountStatsData)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		logger.Error("GetAccountStatsFromDB 에서 에러발생: %s", err)
		// 없는 경우는 회원가입 후 처음 로그인 상황 초기화해줘야 됨.
	} else {
		// accountStats 정보가 db에 있는 경우에만 cache로 가져오기
		cache.CacheAccountStatsInRedis(ctx, uuid, accountStatsData)
	}

	return accountStatsData, err
}

// --- 쓰기 작업은 다음 계층으로 전달하고 캐시를 업데이트/무효화 합니다. ---

func (r *accountCacheRepository) AddAccountRecord(ctx context.Context, uuid []byte, username string, key, salt []byte) error {
	// 쓰기 작업은 DB에 먼저 적용 (새로운 유저 추가)
	return r.next.AddAccountRecord(ctx, uuid, username, key, salt)
}

func (r *accountCacheRepository) UpdateAccountPassword(ctx context.Context, uuid, key, salt []byte) error {
	err := r.next.UpdateAccountPassword(ctx, uuid, key, salt)
	if err != nil {
		return err
	}
	// 캐시 무효화
	redisKey := "session:" + base64.StdEncoding.EncodeToString(uuid)
	return r.cache.Del(ctx, redisKey).Err()
}

// --- 대부분의 읽기 작업은 캐시를 먼저 시도하고, 실패 시 DB로 넘어갑니다. ---

func (r *accountCacheRepository) FetchUsernameBySessionToken(ctx context.Context, token []byte) (string, error) {
	key := "token:" + base64.StdEncoding.EncodeToString(token)
	uuid, err := r.cache.Get(ctx, key).Bytes()

	// 토큰은 Cache에서만 관리, cache에 없으면 err가 반환.
	if err != nil {
		return "", err // Cache miss
	}

	// 캐시에 uuid가 있으면, account 정보에서 username을 찾음
	account, err := r.GetAccount(ctx, uuid)
	username := account.Username

	// 없으면 db에서 찾기
	if err != nil {
		username, err = r.FetchUsernameFromUUID(ctx, uuid)
	}

	return username, nil
}

// --- 나머지 메서드들은 간단하게 다음 계층으로 전달하는 pass-through로 구현합니다. ---
// (추후 필요에 따라 캐싱 로직 추가 가능)

func (r *accountCacheRepository) AddAccountSession(ctx context.Context, username string, token []byte) error {

	// 토큰은 Cache에서만 관리

	// db에서 uuid 가져오기
	uuid, err := r.FetchUUIDFromUsername(ctx, username)
	if err != nil {
		return err
	}

	// uuid와토큰으로 Cache 추가
	// token / uuid
	err = cache.StoreSessionToken(ctx, uuid, token)
	if err != nil {
		return err
	}

	// 유저가 로그인한 것이기 때문에 Cache에 Userdata가 있는지 확인, 초기화
	//err = cache.IsValidCacheData(ctx, uuid)

	// 데이터가 없는 경우 넘어가기
	// if err != nil {
	// 	if !errors.Is(err, redis.Nil) {
	// 		// 데이터가 있으면 return
	// 		return err
	// 	}
	// }

	// db에도 저장쓰
	//r.next.AddAccountSession(ctx, username, token)

	// cache에 해당 uuid를 가진 데이터가 없는 경우
	// db에서 로그인 유저 정보 가져와서 cache에 저장
	// uuid / userData
	// accountData, err := db.GetAccount(uuid)

	// if err != nil {
	// 	// 없는 경우에는 회원이 아닌 거임.
	// 	return err
	// } else {
	// 	// 정보가 있는 경우에만 account 정보 cache로 가져오기
	// 	cache.CacheAccountInRedis(ctx, accountData)
	// }

	// // db에서 로그인 유저 통계 정보 가져와서 cache에 저장
	// accountStatsData, err := db.GetAccountStats(uuid)
	// logger.Info("Login account StatsData : %s", accountStatsData)

	// if err != nil && errors.Is(err, sql.ErrNoRows) {
	// 	logger.Error("GetAccountStatsFromDB 에서 에러발생: %s", err)
	// } else {
	// 	// accountStats 정보가 db에 있는 경우에만 cache로 가져오기
	// 	cache.CacheAccountStatsInRedis(ctx, uuid, accountStatsData)
	// }

	return err
}

func (r *accountCacheRepository) AddDiscordIdByUsername(ctx context.Context, discordId string, username string) error {
	return r.next.AddDiscordIdByUsername(ctx, discordId, username)
}

func (r *accountCacheRepository) AddGoogleIdByUsername(ctx context.Context, googleId string, username string) error {
	return r.next.AddGoogleIdByUsername(ctx, googleId, username)
}

func (r *accountCacheRepository) AddGoogleIdByUUID(ctx context.Context, googleId string, uuid []byte) error {
	return r.next.AddGoogleIdByUUID(ctx, googleId, uuid)
}

func (r *accountCacheRepository) AddDiscordIdByUUID(ctx context.Context, discordId string, uuid []byte) error {
	return r.next.AddDiscordIdByUUID(ctx, discordId, uuid)
}

func (r *accountCacheRepository) FetchUsernameByDiscordId(ctx context.Context, discordId string) (string, error) {
	// This could be cached, but for now, pass through
	return r.next.FetchUsernameByDiscordId(ctx, discordId)
}

func (r *accountCacheRepository) FetchUsernameByGoogleId(ctx context.Context, googleId string) (string, error) {
	return r.next.FetchUsernameByGoogleId(ctx, googleId)
}

func (r *accountCacheRepository) FetchDiscordIdByUsername(ctx context.Context, username string) (string, error) {
	return r.next.FetchDiscordIdByUsername(ctx, username)
}

func (r *accountCacheRepository) FetchGoogleIdByUsername(ctx context.Context, username string) (string, error) {
	return r.next.FetchGoogleIdByUsername(ctx, username)
}

func (r *accountCacheRepository) FetchDiscordIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	return r.next.FetchDiscordIdByUUID(ctx, uuid)
}

func (r *accountCacheRepository) FetchGoogleIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	return r.next.FetchGoogleIdByUUID(ctx, uuid)
}

func (r *accountCacheRepository) CheckUsernameExists(ctx context.Context, username string) (string, error) {
	return r.next.CheckUsernameExists(ctx, username)
}

func (r *accountCacheRepository) FetchLastLoggedInDateByUsername(ctx context.Context, username string) (string, error) {
	return r.next.FetchLastLoggedInDateByUsername(ctx, username)
}

func (r *accountCacheRepository) FetchAdminDetailsByUsername(ctx context.Context, dbUsername string) (AdminSearchResponse, error) {
	return r.next.FetchAdminDetailsByUsername(ctx, dbUsername)
}

func (r *accountCacheRepository) UpdateAccountLastActivity(ctx context.Context, uuid []byte) error {
	// 정보 업데이트는 캐시에 쓰기
	return cache.UpdateAccountLastActivity(ctx, uuid)
	//return r.next.UpdateAccountLastActivity(ctx, uuid)
}

func (r *accountCacheRepository) UpdateAccountStats(ctx context.Context, uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {
	return cache.UpdateAccountStats(ctx, uuid, stats, voucherCounts)
	//return r.next.UpdateAccountStats(ctx, uuid, stats, voucherCounts)
}

func (r *accountCacheRepository) SetAccountBanned(ctx context.Context, uuid []byte, banned bool) error {
	// 밴 계정 데이터 지우기
	cache.DeleteCachedata(ctx, uuid)
	return r.next.SetAccountBanned(ctx, uuid, banned)
}

func (r *accountCacheRepository) FetchAccountKeySaltFromUsername(ctx context.Context, username string) ([]byte, error) {
	return r.next.FetchAccountKeySaltFromUsername(ctx, username)
}

func (r *accountCacheRepository) FetchTrainerIds(ctx context.Context, uuid []byte) (trainerId, secretId int, err error) {
	trainerId, secretId, err = cache.FetchTrainerIds(ctx, uuid)
	if err != nil {
		trainerId, secretId, err = r.next.FetchTrainerIds(ctx, uuid)
	}
	return trainerId, secretId, err
}

func (r *accountCacheRepository) UpdateTrainerIds(ctx context.Context, trainerId, secretId int, uuid []byte) error {
	return cache.UpdateTrainerIds(ctx, trainerId, secretId, uuid)
	//return r.next.UpdateTrainerIds(ctx, trainerId, secretId, uuid)
}

func (r *accountCacheRepository) IsActiveSession(ctx context.Context, uuid []byte, sessionId string) (bool, error) {
	return cache.IsActiveSession(ctx, uuid, sessionId)
	//return r.next.IsActiveSession(ctx, uuid, sessionId)
}

func (r *accountCacheRepository) UpdateActiveSession(ctx context.Context, uuid []byte, clientSessionId string) error {
	return cache.UpdateActiveSession(ctx, uuid, clientSessionId)
	//return r.next.UpdateActiveSession(ctx, uuid, clientSessionId)
}

// Token은 무조건 Cache에서 사용 중
func (r *accountCacheRepository) FetchUUIDFromToken(ctx context.Context, token []byte) ([]byte, error) {
	// 따로 건들이지는 않음.
	return r.next.FetchUUIDFromToken(ctx, token)
}

func (r *accountCacheRepository) RemoveSessionFromToken(ctx context.Context, token []byte) error {
	return cache.RemoveSessionFromToken(ctx, token)
	//return r.next.RemoveSessionFromToken(ctx, token)
}

func (r *accountCacheRepository) FetchUsernameFromUUID(ctx context.Context, uuid []byte) (string, error) {
	// cache에서 안가져옴?
	return r.next.FetchUsernameFromUUID(ctx, uuid)
}

func (r *accountCacheRepository) FetchUUIDFromUsername(ctx context.Context, username string) ([]byte, error) {
	// cache에서 안가져옴?
	return r.next.FetchUUIDFromUsername(ctx, username)
}

func (r *accountCacheRepository) RemoveDiscordIdByUUID(ctx context.Context, uuid []byte) error {
	return r.next.RemoveDiscordIdByUUID(ctx, uuid)
}

func (r *accountCacheRepository) RemoveGoogleIdByUUID(ctx context.Context, uuid []byte) error {
	return r.next.RemoveGoogleIdByUUID(ctx, uuid)
}

func (r *accountCacheRepository) RemoveGoogleIdByUsername(ctx context.Context, username string) error {
	return r.next.RemoveGoogleIdByUsername(ctx, username)
}

func (r *accountCacheRepository) RemoveDiscordIdByUsername(ctx context.Context, username string) error {
	return r.next.RemoveDiscordIdByUsername(ctx, username)
}

func (r *accountCacheRepository) RemoveDiscordIdByDiscordId(ctx context.Context, discordId string) error {
	return r.next.RemoveDiscordIdByDiscordId(ctx, discordId)
}

func (r *accountCacheRepository) RemoveGoogleIdByDiscordId(ctx context.Context, discordId string) error {
	return r.next.RemoveGoogleIdByDiscordId(ctx, discordId)
}
