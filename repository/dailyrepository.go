package repository

import (
	"context"
	"database/sql"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// dailyRepository 구조체 선언 (DB 핸들 포함)
type dailyRepository struct {
	db *sql.DB
}

// 생성자 함수: DB 핸들 주입
func NewDailyRepository(db *sql.DB) *dailyRepository {
	return &dailyRepository{db: db}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *dailyRepository) TryAddDailyRun(ctx context.Context, seed string) (string, error) {
	return db.TryAddDailyRun(seed)
}

func (r *dailyRepository) GetDailyRunSeed(ctx context.Context) (string, error) {
	return db.GetDailyRunSeed()
}

func (r *dailyRepository) AddOrUpdateAccountDailyRun(ctx context.Context, uuid []byte, score int, wave int) error {
	return db.AddOrUpdateAccountDailyRun(uuid, score, wave)
}

func (r *dailyRepository) FetchRankings(ctx context.Context, category int, page int) ([]defs.DailyRanking, error) {
	return db.FetchRankings(category, page)
}

func (r *dailyRepository) FetchRankingPageCount(ctx context.Context, category int) (int, error) {
	return db.FetchRankingPageCount(category)
}
