package defs

import (
	"database/sql"
	"time"
)

// MariaDB 'accounts' 테이블의 한 행을 스캔하기 위한 구조체
type AccountDBRow struct {
	UUID         []byte         // binary(16)
	Username     string         // varchar(16)
	Hash         []byte         // binary(32)
	Salt         []byte         // binary(16)
	Registered   time.Time      // timestamp
	LastLoggedIn sql.NullTime   // timestamp, NULLable
	LastActivity sql.NullTime   // timestamp, NULLable
	Banned       bool           // tinyint(1)
	TrainerID    sql.NullInt32  // smallint(5) unsigned, NULLable
	SecretID     sql.NullInt32  // smallint(5) unsigned, NULLable
	DiscordID    sql.NullString // varchar(32), NULLable
	GoogleID     sql.NullString // varchar(32), NULLable
}

type AccountStatsData struct {
	UUID                  []byte `db:"uuid"` // DB에서 읽어올 때 []byte, 실제 사용 시 문자열로 변환 가능
	PlayTime              int    `db:"playTime"`
	Battles               int    `db:"battles"`
	ClassicSessionsPlayed int    `db:"classicSessionsPlayed"`
	SessionsWon           int    `db:"sessionsWon"`
	HighestEndlessWave    int    `db:"highestEndlessWave"`
	HighestLevel          int    `db:"highestLevel"`
	PokemonSeen           int    `db:"pokemonSeen"`
	PokemonDefeated       int    `db:"pokemonDefeated"`
	PokemonCaught         int    `db:"pokemonCaught"`
	PokemonHatched        int    `db:"pokemonHatched"`
	EggsPulled            int    `db:"eggsPulled"`
	RegularVouchers       int    `db:"regularVouchers"`
	PlusVouchers          int    `db:"plusVouchers"`
	PremiumVouchers       int    `db:"premiumVouchers"`
	GoldenVouchers        int    `db:"goldenVouchers"`
}

// Redis에 JSON으로 저장될 데이터 구조체 (uuid는 키로 사용되므로 여기엔 없음)
type AccountRedisData struct {
	Username     string     `json:"username"`
	Hash         string     `json:"hash"` // Hex-encoded string
	Salt         string     `json:"salt"` // Hex-encoded string
	Registered   time.Time  `json:"registered"`
	LastLoggedIn *time.Time `json:"lastLoggedIn,omitempty"`
	LastActivity *time.Time `json:"lastActivity,omitempty"`
	Banned       bool       `json:"banned"`
	TrainerID    *uint16    `json:"trainerId,omitempty"`
	SecretID     *uint16    `json:"secretId,omitempty"`
	DiscordID    *string    `json:"discordId,omitempty"`
	GoogleID     *string    `json:"googleId,omitempty"`
}

// Redis에 저장될 accountStats 데이터 (UUID 제외)
// 이 구조체를 사용하는 것이 더 명시적이고 깔끔합니다.
type AccountStatsRedisData struct {
	PlayTime              int `json:"playTime"`
	Battles               int `json:"battles"`
	ClassicSessionsPlayed int `json:"classicSessionsPlayed"`
	SessionsWon           int `json:"sessionsWon"`
	HighestEndlessWave    int `json:"highestEndlessWave"`
	HighestLevel          int `json:"highestLevel"`
	PokemonSeen           int `json:"pokemonSeen"`
	PokemonDefeated       int `json:"pokemonDefeated"`
	PokemonCaught         int `json:"pokemonCaught"`
	PokemonHatched        int `json:"pokemonHatched"`
	EggsPulled            int `json:"eggsPulled"`
	RegularVouchers       int `json:"regularVouchers"`
	PlusVouchers          int `json:"plusVouchers"`
	PremiumVouchers       int `json:"premiumVouchers"`
	GoldenVouchers        int `json:"goldenVouchers"`
}
