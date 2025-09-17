package repository

import (
	"context"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

// dailyRepository 구조체 선언 (DB 핸들 포함)
type dailyRepository struct {
	cache *redis.Client
	next  DailyRepository // 다음 계층 (DB 레포지토리)
}

// 생성자 함수: DB 핸들 주입
func NewDailyRepository(redisClient *redis.Client, nextRepo DailyRepository) DailyRepository {
	return &dailyRepository{cache: redisClient, next: nextRepo}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *dailyRepository) TryAddDailyRun(ctx context.Context, seed string) (string, error) {
	return r.next.TryAddDailyRun(ctx, seed)
}

func (r *dailyRepository) GetDailyRunSeed(ctx context.Context) (string, error) {
	return r.next.GetDailyRunSeed(ctx)
}

func (r *dailyRepository) AddOrUpdateAccountDailyRun(ctx context.Context, uuid []byte, score int, wave int) error {
	return r.next.AddOrUpdateAccountDailyRun(ctx, uuid, score, wave)
}

func (r *dailyRepository) FetchRankings(ctx context.Context, category int, page int) ([]defs.DailyRanking, error) {
	return r.next.FetchRankings(ctx, category, page)
}

func (r *dailyRepository) FetchRankingPageCount(ctx context.Context, category int) (int, error) {
	return r.next.FetchRankingPageCount(ctx, category)
}
