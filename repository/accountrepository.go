package repository

import (
	"context"

	"github.com/pagefaultgames/rogueserver/defs"
)

// accountRepository 구조체 선언
type accountRepository struct {
	db    AccountRepository
	cache AccountRepository
}

func NewAccountRepository(db AccountRepository, cache AccountRepository) *accountRepository {
	return &accountRepository{db: db, cache: cache}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *accountRepository) GetAccount(ctx context.Context, uuid []byte) (defs.AccountDBRow, error) {
	return r.db.GetAccount(ctx, uuid)
}

func (r *accountRepository) GetAccountStats(ctx context.Context, uuid []byte) (defs.AccountStatsData, error) {
	return r.db.GetAccountStats(ctx, uuid)
}

func (r *accountRepository) AddAccountRecord(ctx context.Context, uuid []byte, username string, key, salt []byte) error {
	return r.db.AddAccountRecord(ctx, uuid, username, key, salt)
}

func (r *accountRepository) AddAccountSession(ctx context.Context, username string, token []byte) error {
	return r.db.AddAccountSession(ctx, username, token)
}

func (r *accountRepository) AddDiscordIdByUsername(ctx context.Context, discordId string, username string) error {
	return r.db.AddDiscordIdByUsername(ctx, discordId, username)
}

func (r *accountRepository) AddGoogleIdByUsername(ctx context.Context, googleId string, username string) error {
	return r.db.AddGoogleIdByUsername(ctx, googleId, username)
}

func (r *accountRepository) AddGoogleIdByUUID(ctx context.Context, googleId string, uuid []byte) error {
	return r.db.AddGoogleIdByUUID(ctx, googleId, uuid)
}

func (r *accountRepository) AddDiscordIdByUUID(ctx context.Context, discordId string, uuid []byte) error {
	return r.db.AddDiscordIdByUUID(ctx, discordId, uuid)
}

func (r *accountRepository) FetchUsernameByDiscordId(ctx context.Context, discordId string) (string, error) {
	return r.db.FetchUsernameByDiscordId(ctx, discordId)
}

func (r *accountRepository) FetchUsernameByGoogleId(ctx context.Context, googleId string) (string, error) {
	return r.db.FetchUsernameByGoogleId(ctx, googleId)
}

func (r *accountRepository) FetchDiscordIdByUsername(ctx context.Context, username string) (string, error) {
	return r.db.FetchDiscordIdByUsername(ctx, username)
}

func (r *accountRepository) FetchGoogleIdByUsername(ctx context.Context, username string) (string, error) {
	return r.db.FetchGoogleIdByUsername(ctx, username)
}

func (r *accountRepository) FetchDiscordIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	return r.db.FetchDiscordIdByUUID(ctx, uuid)
}

func (r *accountRepository) FetchGoogleIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	return r.db.FetchGoogleIdByUUID(ctx, uuid)
}

func (r *accountRepository) FetchUsernameBySessionToken(ctx context.Context, token []byte) (string, error) {
	return r.db.FetchUsernameBySessionToken(ctx, token)
}

func (r *accountRepository) CheckUsernameExists(ctx context.Context, username string) (string, error) {
	return r.db.CheckUsernameExists(ctx, username)
}

func (r *accountRepository) FetchLastLoggedInDateByUsername(ctx context.Context, username string) (string, error) {
	return r.db.FetchLastLoggedInDateByUsername(ctx, username)
}

func (r *accountRepository) FetchAdminDetailsByUsername(ctx context.Context, dbUsername string) (AdminSearchResponse, error) {
	return r.db.FetchAdminDetailsByUsername(ctx, dbUsername)
}

func (r *accountRepository) UpdateAccountPassword(ctx context.Context, uuid, key, salt []byte) error {
	return r.db.UpdateAccountPassword(ctx, uuid, key, salt)
}

func (r *accountRepository) UpdateAccountLastActivity(ctx context.Context, uuid []byte) error {
	return r.db.UpdateAccountLastActivity(ctx, uuid)
}

func (r *accountRepository) UpdateAccountStats(ctx context.Context, uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {
	return r.db.UpdateAccountStats(ctx, uuid, stats, voucherCounts)
}

func (r *accountRepository) SetAccountBanned(ctx context.Context, uuid []byte, banned bool) error {
	return r.db.SetAccountBanned(ctx, uuid, banned)
}

func (r *accountRepository) FetchAccountKeySaltFromUsername(ctx context.Context, username string) ([]byte, error) {
	return r.db.FetchAccountKeySaltFromUsername(ctx, username)
}

func (r *accountRepository) FetchTrainerIds(ctx context.Context, uuid []byte) (trainerId, secretId int, err error) {
	return r.db.FetchTrainerIds(ctx, uuid)
}

func (r *accountRepository) UpdateTrainerIds(ctx context.Context, trainerId, secretId int, uuid []byte) error {
	return r.db.UpdateTrainerIds(ctx, trainerId, secretId, uuid)
}

func (r *accountRepository) IsActiveSession(ctx context.Context, uuid []byte, sessionId string) (bool, error) {
	return r.db.IsActiveSession(ctx, uuid, sessionId)
}

func (r *accountRepository) UpdateActiveSession(ctx context.Context, uuid []byte, clientSessionId string) error {
	return r.db.UpdateActiveSession(ctx, uuid, clientSessionId)
}

func (r *accountRepository) FetchUUIDFromToken(ctx context.Context, token []byte) ([]byte, error) {
	return r.db.FetchUUIDFromToken(ctx, token)
}

func (r *accountRepository) RemoveSessionFromToken(ctx context.Context, token []byte) error {
	return r.db.RemoveSessionFromToken(ctx, token)
}

func (r *accountRepository) FetchUsernameFromUUID(ctx context.Context, uuid []byte) (string, error) {
	return r.db.FetchUsernameFromUUID(ctx, uuid)
}

func (r *accountRepository) FetchUUIDFromUsername(ctx context.Context, username string) ([]byte, error) {
	return r.db.FetchUUIDFromUsername(ctx, username)
}

func (r *accountRepository) RemoveDiscordIdByUUID(ctx context.Context, uuid []byte) error {
	return r.db.RemoveDiscordIdByUUID(ctx, uuid)
}

func (r *accountRepository) RemoveGoogleIdByUUID(ctx context.Context, uuid []byte) error {
	return r.db.RemoveGoogleIdByUUID(ctx, uuid)
}

func (r *accountRepository) RemoveGoogleIdByUsername(ctx context.Context, username string) error {
	return r.db.RemoveGoogleIdByUsername(ctx, username)
}

func (r *accountRepository) RemoveDiscordIdByUsername(ctx context.Context, username string) error {
	return r.db.RemoveDiscordIdByUsername(ctx, username)
}

func (r *accountRepository) RemoveDiscordIdByDiscordId(ctx context.Context, discordId string) error {
	return r.db.RemoveDiscordIdByDiscordId(ctx, discordId)
}

func (r *accountRepository) RemoveGoogleIdByDiscordId(ctx context.Context, discordId string) error {
	return r.db.RemoveGoogleIdByDiscordId(ctx, discordId)
}
