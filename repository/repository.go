package repository

import (
	"context"

	"github.com/pagefaultgames/rogueserver/defs"
)

type Repositories struct {
	Account  AccountRepository
	Daily    DailyRepository
	Game     GameRepository
	Savedata SavedataRepository
}

type AccountRepository interface {
	GetAccount(ctx context.Context, uuid []byte) (defs.AccountDBRow, error)
	GetAccountStats(ctx context.Context, uuid []byte) (defs.AccountStatsData, error)
	AddAccountRecord(ctx context.Context, uuid []byte, username string, key, salt []byte) error
	AddAccountSession(ctx context.Context, username string, token []byte) error

	AddDiscordIdByUsername(ctx context.Context, discordId string, username string) error
	AddGoogleIdByUsername(ctx context.Context, googleId string, username string) error
	AddGoogleIdByUUID(ctx context.Context, googleId string, uuid []byte) error
	AddDiscordIdByUUID(ctx context.Context, discordId string, uuid []byte) error

	FetchUsernameByDiscordId(ctx context.Context, discordId string) (string, error)
	FetchUsernameByGoogleId(ctx context.Context, googleId string) (string, error)
	FetchDiscordIdByUsername(ctx context.Context, username string) (string, error)
	FetchGoogleIdByUsername(ctx context.Context, username string) (string, error)
	FetchDiscordIdByUUID(ctx context.Context, uuid []byte) (string, error)
	FetchGoogleIdByUUID(ctx context.Context, uuid []byte) (string, error)
	FetchUsernameBySessionToken(ctx context.Context, token []byte) (string, error)
	CheckUsernameExists(ctx context.Context, username string) (string, error)
	FetchLastLoggedInDateByUsername(ctx context.Context, username string) (string, error)
	FetchAdminDetailsByUsername(ctx context.Context, dbUsername string) (AdminSearchResponse, error)

	UpdateAccountPassword(ctx context.Context, uuid, key, salt []byte) error
	UpdateAccountLastActivity(ctx context.Context, uuid []byte) error
	UpdateAccountStats(ctx context.Context, uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error
	SetAccountBanned(ctx context.Context, uuid []byte, banned bool) error
	FetchAccountKeySaltFromUsername(ctx context.Context, username string) ([]byte, error)
	FetchTrainerIds(ctx context.Context, uuid []byte) (trainerId, secretId int, err error)
	UpdateTrainerIds(ctx context.Context, trainerId, secretId int, uuid []byte) error

	IsActiveSession(ctx context.Context, uuid []byte, sessionId string) (bool, error)
	UpdateActiveSession(ctx context.Context, uuid []byte, clientSessionId string) error
	FetchUUIDFromToken(ctx context.Context, token []byte) ([]byte, error)
	RemoveSessionFromToken(ctx context.Context, token []byte) error
	FetchUsernameFromUUID(ctx context.Context, uuid []byte) (string, error)
	FetchUUIDFromUsername(ctx context.Context, username string) ([]byte, error)

	RemoveDiscordIdByUUID(ctx context.Context, uuid []byte) error
	RemoveGoogleIdByUUID(ctx context.Context, uuid []byte) error
	RemoveGoogleIdByUsername(ctx context.Context, username string) error
	RemoveDiscordIdByUsername(ctx context.Context, username string) error
	RemoveDiscordIdByDiscordId(ctx context.Context, discordId string) error
	RemoveGoogleIdByDiscordId(ctx context.Context, discordId string) error
}

type AdminSearchResponse struct {
	Username     string `json:"username"`
	DiscordId    string `json:"discordId"`
	GoogleId     string `json:"googleId"`
	LastActivity string `json:"lastLoggedIn"` // TODO: this is currently lastLoggedIn to match server PR #54 with pokerogue PR #4198. We're hotfixing the server with this PR to return lastActivity, but we're not hotfixing the client, so are leaving this as lastLoggedIn so that it still talks to the client properly
	Registered   string `json:"registered"`
}

type DailyRepository interface {
	TryAddDailyRun(ctx context.Context, seed string) (string, error)
	GetDailyRunSeed(ctx context.Context) (string, error)
	AddOrUpdateAccountDailyRun(ctx context.Context, uuid []byte, score int, wave int) error
	FetchRankings(ctx context.Context, category int, page int) ([]defs.DailyRanking, error)
	FetchRankingPageCount(ctx context.Context, category int) (int, error)
}

type GameRepository interface {
	FetchPlayerCount(ctx context.Context) (int, error)
	FetchBattleCount(ctx context.Context) (int, error)
	FetchClassicSessionCount(ctx context.Context) (int, error)
}

type SavedataRepository interface {
	TryAddSeedCompletion(ctx context.Context, uuid []byte, seed string, mode int) (bool, error)
	ReadSeedCompleted(ctx context.Context, uuid []byte, seed string) (bool, error)

	ReadSystemSaveData(ctx context.Context, uuid []byte) (defs.SystemSaveData, error)
	StoreSystemSaveData(ctx context.Context, uuid []byte, data defs.SystemSaveData) error
	StoreSystemSaveDataS3(ctx context.Context, uuid []byte, data defs.SystemSaveData) error
	DeleteSystemSaveData(ctx context.Context, uuid []byte) error

	ReadSessionSaveData(ctx context.Context, uuid []byte, slot int) (defs.SessionSaveData, error)
	GetLatestSessionSaveDataSlot(ctx context.Context, uuid []byte) (int, error)
	StoreSessionSaveData(ctx context.Context, uuid []byte, data defs.SessionSaveData, slot int) error
	DeleteSessionSaveData(ctx context.Context, uuid []byte, slot int) error

	StoreSessionSaveDataBulk(ctx context.Context, uuids [][]byte, sessionsDataMapList []defs.SessionSaveData, slot int) error
	StoreSystemSaveDataBulk(ctx context.Context, uuids [][]byte, systemsDataList []defs.SystemSaveData) error

	RetrievePlaytime(ctx context.Context, uuid []byte) (int, error)
	GetSystemSaveFromS3(ctx context.Context, uuid []byte) (defs.SystemSaveData, error)
}
