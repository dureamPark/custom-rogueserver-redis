package dbrepository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/repository"
)

// accountRepository는 AccountRepository 인터페이스의 데이터베이스 구현체입니다.
type accountDBRepository struct {
	db *sql.DB
}

// NewAccountRepository는 데이터베이스와 통신하는 새로운 account repository를 생성합니다.
func NewAccountDBRepository(db *sql.DB) repository.AccountRepository {
	return &accountDBRepository{db: db}
}

func (r *accountDBRepository) GetAccount(ctx context.Context, uuid []byte) (defs.AccountDBRow, error) {
	var account defs.AccountDBRow

	query := `
		SELECT uuid, username, hash, salt, registered, lastLoggedIn, 
		       lastActivity, banned, trainerId, secretId, discordId, googleId 
		FROM accounts 
		WHERE uuid = ?
	`

	row := r.db.QueryRowContext(ctx, query, uuid)
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
			return account, fmt.Errorf("UUID '%s'에 해당하는 계정을 찾을 수 없음: %w", uuid, err)
		}
		return account, fmt.Errorf("DB에서 계정 조회 중 오류 (UUID: %s): %w", uuid, err)
	}

	return account, nil
}

func (r *accountDBRepository) GetAccountStats(ctx context.Context, uuidBytes []byte) (defs.AccountStatsData, error) {
	var stats defs.AccountStatsData

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

	row := r.db.QueryRowContext(ctx, query, uuidBytes)
	err := row.Scan(
		&stats.UUID,
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
			return stats, fmt.Errorf("UUID '%s'에 해당하는 통계 데이터를 찾을 수 없음: %w", base64.StdEncoding.EncodeToString(uuidBytes), err)
		}
		return stats, fmt.Errorf("DB에서 통계 데이터 조회 중 오류 (UUID: %s): %w", base64.StdEncoding.EncodeToString(uuidBytes), err)
	}

	return stats, nil
}

func (r *accountDBRepository) AddAccountRecord(ctx context.Context, uuid []byte, username string, key, salt []byte) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO accounts (uuid, username, hash, salt, registered) VALUES (?, ?, ?, ?, UTC_TIMESTAMP())", uuid, username, key, salt)
	return err
}

func (r *accountDBRepository) AddAccountSession(ctx context.Context, username string, token []byte) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO sessions (uuid, token, expire) SELECT a.uuid, ?, DATE_ADD(UTC_TIMESTAMP(), INTERVAL 1 WEEK) FROM accounts a WHERE a.username = ?", token, username)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, "UPDATE accounts SET lastLoggedIn = UTC_TIMESTAMP() WHERE username = ?", username)
	return err
}

func (r *accountDBRepository) AddDiscordIdByUsername(ctx context.Context, discordId string, username string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET discordId = ? WHERE username = ?", discordId, username)
	return err
}

func (r *accountDBRepository) AddGoogleIdByUsername(ctx context.Context, googleId string, username string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET googleId = ? WHERE username = ?", googleId, username)
	return err
}

func (r *accountDBRepository) AddGoogleIdByUUID(ctx context.Context, googleId string, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET googleId = ? WHERE uuid = ?", googleId, uuid)
	return err
}

func (r *accountDBRepository) AddDiscordIdByUUID(ctx context.Context, discordId string, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET discordId = ? WHERE uuid = ?", discordId, uuid)
	return err
}

func (r *accountDBRepository) FetchUsernameByDiscordId(ctx context.Context, discordId string) (string, error) {
	var username string
	err := r.db.QueryRowContext(ctx, "SELECT username FROM accounts WHERE discordId = ?", discordId).Scan(&username)
	return username, err
}

func (r *accountDBRepository) FetchUsernameByGoogleId(ctx context.Context, googleId string) (string, error) {
	var username string
	err := r.db.QueryRowContext(ctx, "SELECT username FROM accounts WHERE googleId = ?", googleId).Scan(&username)
	return username, err
}

func (r *accountDBRepository) FetchDiscordIdByUsername(ctx context.Context, username string) (string, error) {
	var discordId sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT discordId FROM accounts WHERE username = ?", username).Scan(&discordId)
	if err != nil {
		return "", err
	}
	if !discordId.Valid {
		return "", nil
	}
	return discordId.String, nil
}

func (r *accountDBRepository) FetchGoogleIdByUsername(ctx context.Context, username string) (string, error) {
	var googleId sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT googleId FROM accounts WHERE username = ?", username).Scan(&googleId)
	if err != nil {
		return "", err
	}
	if !googleId.Valid {
		return "", nil
	}
	return googleId.String, nil
}

func (r *accountDBRepository) FetchDiscordIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	var discordId sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT discordId FROM accounts WHERE uuid = ?", uuid).Scan(&discordId)
	if err != nil {
		return "", err
	}
	if !discordId.Valid {
		return "", nil
	}
	return discordId.String, nil
}

func (r *accountDBRepository) FetchGoogleIdByUUID(ctx context.Context, uuid []byte) (string, error) {
	var googleId sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT googleId FROM accounts WHERE uuid = ?", uuid).Scan(&googleId)
	if err != nil {
		return "", err
	}
	if !googleId.Valid {
		return "", nil
	}
	return googleId.String, nil
}

func (r *accountDBRepository) FetchUsernameBySessionToken(ctx context.Context, token []byte) (string, error) {
	var username string
	err := r.db.QueryRowContext(ctx, "SELECT a.username FROM accounts a JOIN sessions s ON a.uuid = s.uuid WHERE s.token = ?", token).Scan(&username)
	return username, err
}

func (r *accountDBRepository) CheckUsernameExists(ctx context.Context, username string) (string, error) {
	var dbUsername sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT username FROM accounts WHERE username = ?", username).Scan(&dbUsername)
	if err != nil {
		return "", err
	}
	if !dbUsername.Valid {
		return "", nil
	}
	return dbUsername.String, nil
}

func (r *accountDBRepository) FetchLastLoggedInDateByUsername(ctx context.Context, username string) (string, error) {
	var lastLoggedIn sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT lastLoggedIn FROM accounts WHERE username = ?", username).Scan(&lastLoggedIn)
	if err != nil {
		return "", err
	}
	if !lastLoggedIn.Valid {
		return "", nil
	}
	return lastLoggedIn.String, nil
}

func (r *accountDBRepository) FetchAdminDetailsByUsername(ctx context.Context, dbUsername string) (repository.AdminSearchResponse, error) {
	var username, discordId, googleId, lastActivity, registered sql.NullString
	var adminResponse repository.AdminSearchResponse

	err := r.db.QueryRowContext(ctx, "SELECT username, discordId, googleId, lastActivity, registered from accounts WHERE username = ?", dbUsername).Scan(&username, &discordId, &googleId, &lastActivity, &registered)
	if err != nil {
		return adminResponse, err
	}

	adminResponse = repository.AdminSearchResponse{
		Username:     username.String,
		DiscordId:    discordId.String,
		GoogleId:     googleId.String,
		LastActivity: lastActivity.String,
		Registered:   registered.String,
	}

	return adminResponse, nil
}

func (r *accountDBRepository) UpdateAccountPassword(ctx context.Context, uuid, key, salt []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET hash = ?, salt = ? WHERE uuid = ?", key, salt, uuid)
	return err
}

func (r *accountDBRepository) UpdateAccountLastActivity(ctx context.Context, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET lastActivity = UTC_TIMESTAMP() WHERE uuid = ?", uuid)
	return err
}

func (r *accountDBRepository) UpdateAccountStats(ctx context.Context, uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {
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

	_, err := r.db.ExecContext(ctx, query, statArgs...)
	return err
}

func (r *accountDBRepository) SetAccountBanned(ctx context.Context, uuid []byte, banned bool) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET banned = ? WHERE uuid = ?", banned, uuid)
	return err
}

func (r *accountDBRepository) FetchAccountKeySaltFromUsername(ctx context.Context, username string) ([]byte, error) {
	var key, salt []byte
	err := r.db.QueryRowContext(ctx, "SELECT hash, salt FROM accounts WHERE username = ?", username).Scan(&key, &salt)
	return key, err
}

func (r *accountDBRepository) FetchTrainerIds(ctx context.Context, uuid []byte) (trainerId, secretId int, err error) {
	err = r.db.QueryRowContext(ctx, "SELECT trainerId, secretId FROM accounts WHERE uuid = ?", uuid).Scan(&trainerId, &secretId)
	return trainerId, secretId, err
}

func (r *accountDBRepository) UpdateTrainerIds(ctx context.Context, trainerId, secretId int, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET trainerId = ?, secretId = ? WHERE uuid = ?", trainerId, secretId, uuid)
	return err
}

func (r *accountDBRepository) IsActiveSession(ctx context.Context, uuid []byte, sessionId string) (bool, error) {
	var id string
	err := r.db.QueryRowContext(ctx, "SELECT clientSessionId FROM activeClientSessions WHERE uuid = ?", uuid).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = r.UpdateActiveSession(ctx, uuid, sessionId)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		return false, err
	}

	return id == "" || id == sessionId, nil
}

func (r *accountDBRepository) UpdateActiveSession(ctx context.Context, uuid []byte, clientSessionId string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO activeClientSessions (uuid, clientSessionId) VALUES (?, ?) ON DUPLICATE KEY UPDATE clientSessionId = ?", uuid, clientSessionId, clientSessionId)
	return err
}

func (r *accountDBRepository) FetchUUIDFromToken(ctx context.Context, token []byte) ([]byte, error) {
	var uuid []byte
	err := r.db.QueryRowContext(ctx, "SELECT uuid FROM sessions WHERE token = ?", token).Scan(&uuid)
	return uuid, err
}

func (r *accountDBRepository) RemoveSessionFromToken(ctx context.Context, token []byte) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (r *accountDBRepository) FetchUsernameFromUUID(ctx context.Context, uuid []byte) (string, error) {
	var username string
	err := r.db.QueryRowContext(ctx, "SELECT username FROM accounts WHERE uuid = ?", uuid).Scan(&username)
	return username, err
}

func (r *accountDBRepository) FetchUUIDFromUsername(ctx context.Context, username string) ([]byte, error) {
	var uuid []byte
	err := r.db.QueryRowContext(ctx, "SELECT uuid FROM accounts WHERE username = ?", username).Scan(&uuid)
	return uuid, err
}

func (r *accountDBRepository) RemoveDiscordIdByUUID(ctx context.Context, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET discordId = NULL WHERE uuid = ?", uuid)
	return err
}

func (r *accountDBRepository) RemoveGoogleIdByUUID(ctx context.Context, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET googleId = NULL WHERE uuid = ?", uuid)
	return err
}

func (r *accountDBRepository) RemoveGoogleIdByUsername(ctx context.Context, username string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET googleId = NULL WHERE username = ?", username)
	return err
}

func (r *accountDBRepository) RemoveDiscordIdByUsername(ctx context.Context, username string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET discordId = NULL WHERE username = ?", username)
	return err
}

func (r *accountDBRepository) RemoveDiscordIdByDiscordId(ctx context.Context, discordId string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET discordId = NULL WHERE discordId = ?", discordId)
	return err
}

func (r *accountDBRepository) RemoveGoogleIdByDiscordId(ctx context.Context, discordId string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET googleId = NULL WHERE discordId = ?", discordId)
	return err
}
