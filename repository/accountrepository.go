package repository

import (
	"context"
	"database/sql"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// accountRepository 구조체 선언 (DB 핸들 포함)
type accountRepository struct {
	db *sql.DB
}

// 생성자 함수: DB 핸들 주입
func NewAccountRepository(db *sql.DB) *accountRepository {
	return &accountRepository{db: db}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *accountRepository) GetAccount(ctx context.Context, uuid []byte) (defs.AccountDBRow, error) {
	return db.GetAccountFromDB(uuid)
}

func (r *accountRepository) GetAccountStats(ctx context.Context, uuid []byte) (defs.AccountStatsData, error) {
	return db.GetAccountStatsFromDB(uuid)
}

func (r *accountRepository) AddAccountRecord(ctx context.Context, uuid []byte, username string, key, salt []byte) error {
	return db.AddAccountRecord(uuid, username, key, salt)
}

func (r *accountRepository) AddAccountSession(ctx context.Context, username string, token []byte) error {
	return db.AddAccountSession(username, token)
}

func (r *accountRepository) AddDiscordIdByUsername(ctx context.Context, discordId string, username string) error {
	return db.AddDiscordIdByUsername(discordId, username)
}

func (r *accountRepository) AddGoogleIdByUsername(ctx context.Context, googleId string, username string) error {
	return db.AddGoogleIdByUsername(googleId, username)
}

func (r *accountRepository) AddGoogleIdByUUID(ctx context.Context, googleId string, uuid []byte) error {
	return db.AddGoogleIdByUUID(googleId, uuid)
}

func (r *accountRepository) AddDiscordIdByUUID(ctx context.Context, discordId string, uuid []byte) error {
	return db.AddDiscordIdByUUID(discordId, uuid)
}

func (r *accountRepository) FetchUsernameByDiscordId(ctx context.Context, discordId string) (string, error) {
	return db.FetchUsernameByDiscordId(discordId)
}

func (r *accountRepository) FetchUsernameByGoogleId(ctx context.Context, googleId string) (string, error) {
	return db.FetchUsernameByGoogleId(googleId)
}

func (r *accountRepository) FetchDiscordIdByUsername(ctx context.Context, username string) (string, error) {
	return db.FetchDiscordIdByUsername(username)
}

func (r *accountRepository) FetchGoogleIdByUsername(ctx context.Context, username string) (string, error) {
	return db.FetchGoogleIdByUsername(username)
}

func (r *accountRepository) FetchDiscordIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	return db.FetchDiscordIdByUUID(uuid)
}

func (r *accountRepository) FetchGoogleIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	return db.FetchGoogleIdByUUID(uuid)
}

func (r *accountRepository) FetchUsernameBySessionToken(ctx context.Context, token []byte) (string, error) {
	return db.FetchUsernameBySessionToken(token)
}

func (r *accountRepository) CheckUsernameExists(ctx context.Context, username string) (string, error) {
	return db.CheckUsernameExists(username)
}

func (r *accountRepository) FetchLastLoggedInDateByUsername(ctx context.Context, username string) (string, error) {
	return db.FetchLastLoggedInDateByUsername(username)
}

func (r *accountRepository) FetchAdminDetailsByUsername(ctx context.Context, dbUsername string) (AdminSearchResponse, error) {
	dbResponse, err := db.FetchAdminDetailsByUsername(dbUsername)
	if err != nil {
		return AdminSearchResponse{}, err
	}
	// Convert db.AdminSearchResponse to repository.AdminSearchResponse
	return AdminSearchResponse{
		Username:     dbResponse.Username,
		DiscordId:    dbResponse.DiscordId,
		GoogleId:     dbResponse.GoogleId,
		LastActivity: dbResponse.LastActivity,
		Registered:   dbResponse.Registered,
	}, nil
}

func (r *accountRepository) UpdateAccountPassword(ctx context.Context, uuid, key, salt []byte) error {
	return db.UpdateAccountPassword(uuid, key, salt)
}

func (r *accountRepository) UpdateAccountLastActivity(ctx context.Context, uuid []byte) error {
	return db.UpdateAccountLastActivity(uuid)
}

func (r *accountRepository) UpdateAccountStats(ctx context.Context, uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {
	return db.UpdateAccountStats(uuid, stats, voucherCounts)
}

func (r *accountRepository) SetAccountBanned(ctx context.Context, uuid []byte, banned bool) error {
	return db.SetAccountBanned(uuid, banned)
}

func (r *accountRepository) FetchAccountKeySaltFromUsername(ctx context.Context, username string) ([]byte, error) {
	return db.FetchAccountKeySaltFromUsername(username)
}

func (r *accountRepository) FetchTrainerIds(ctx context.Context, uuid []byte) (trainerId, secretId int, err error) {
	return db.FetchTrainerIds(uuid)
}

func (r *accountRepository) UpdateTrainerIds(ctx context.Context, trainerId, secretId int, uuid []byte) error {
	return db.UpdateTrainerIds(trainerId, secretId, uuid)
}

func (r *accountRepository) IsActiveSession(ctx context.Context, uuid []byte, sessionId string) (bool, error) {
	return db.IsActiveSession(uuid, sessionId)
}

func (r *accountRepository) UpdateActiveSession(ctx context.Context, uuid []byte, clientSessionId string) error {
	return db.UpdateActiveSession(uuid, clientSessionId)
}

func (r *accountRepository) FetchUUIDFromToken(ctx context.Context, token []byte) ([]byte, error) {
	return db.FetchUUIDFromToken(token)
}

func (r *accountRepository) RemoveSessionFromToken(ctx context.Context, token []byte) error {
	return db.RemoveSessionFromToken(token)
}

func (r *accountRepository) FetchUsernameFromUUID(ctx context.Context, uuid []byte) (string, error) {
	return db.FetchUsernameFromUUID(uuid)
}

func (r *accountRepository) FetchUUIDFromUsername(ctx context.Context, username string) ([]byte, error) {
	return db.FetchUUIDFromUsername(username)
}

func (r *accountRepository) RemoveDiscordIdByUUID(ctx context.Context, uuid []byte) error {
	return db.RemoveDiscordIdByUUID(uuid)
}

func (r *accountRepository) RemoveGoogleIdByUUID(ctx context.Context, uuid []byte) error {
	return db.RemoveGoogleIdByUUID(uuid)
}

func (r *accountRepository) RemoveGoogleIdByUsername(ctx context.Context, username string) error {
	return db.RemoveGoogleIdByUsername(username)
}

func (r *accountRepository) RemoveDiscordIdByUsername(ctx context.Context, username string) error {
	return db.RemoveDiscordIdByUsername(username)
}

func (r *accountRepository) RemoveDiscordIdByDiscordId(ctx context.Context, discordId string) error {
	return db.RemoveDiscordIdByDiscordId(discordId)
}

func (r *accountRepository) RemoveGoogleIdByDiscordId(ctx context.Context, discordId string) error {
	return db.RemoveGoogleIdByDiscordId(discordId)
}
