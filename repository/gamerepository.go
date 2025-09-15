package repository

import (
	"context"
	"database/sql"

	"github.com/pagefaultgames/rogueserver/db"
)

// gameRepository 구조체 선언 (DB 핸들 포함)
type gameRepository struct {
	db *sql.DB
}

// 생성자 함수: DB 핸들 주입
func NewGameRepository(db *sql.DB) *gameRepository {
	return &gameRepository{db: db}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *gameRepository) FetchPlayerCount(ctx context.Context) (int, error) {
	return db.FetchPlayerCount()
}

func (r *gameRepository) FetchBattleCount(ctx context.Context) (int, error) {
	return db.FetchBattleCount()
}

func (r *gameRepository) FetchClassicSessionCount(ctx context.Context) (int, error) {
	return db.FetchClassicSessionCount()
}
