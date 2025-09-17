package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// gameRepository 구조체 선언 (DB 핸들 포함)
type gameRepository struct {
	cache *redis.Client
	next  GameRepository
}

// 생성자 함수: DB 핸들 주입
func NewGameRepository(redisClient *redis.Client, nextRepo GameRepository) *gameRepository {
	return &gameRepository{cache: redisClient, next: nextRepo}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *gameRepository) FetchPlayerCount(ctx context.Context) (int, error) {
	return r.next.FetchPlayerCount(ctx)
}

func (r *gameRepository) FetchBattleCount(ctx context.Context) (int, error) {
	return r.next.FetchBattleCount(ctx)
}

func (r *gameRepository) FetchClassicSessionCount(ctx context.Context) (int, error) {
	return r.next.FetchClassicSessionCount(ctx)
}
