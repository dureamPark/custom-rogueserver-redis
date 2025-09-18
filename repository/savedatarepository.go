package repository

import (
	"context"
	"errors"
	"os"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

// savedataRepository 구조체 선언 (DB 핸들 포함)
type savedataRepository struct {
	cache *redis.Client
	next  SavedataRepository
}

// 생성자 함수: DB 핸들 주입
func NewSavedataRepository(redisClient *redis.Client, nextRepo SavedataRepository) SavedataRepository {
	return &savedataRepository{cache: redisClient, next: nextRepo}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *savedataRepository) TryAddSeedCompletion(ctx context.Context, uuid []byte, seed string, mode int) (bool, error) {
	return r.next.TryAddSeedCompletion(ctx, uuid, seed, mode)
}

func (r *savedataRepository) ReadSeedCompleted(ctx context.Context, uuid []byte, seed string) (bool, error) {
	return r.next.ReadSeedCompleted(ctx, uuid, seed)
}

func (r *savedataRepository) ReadSystemSaveData(ctx context.Context, uuid []byte) (defs.SystemSaveData, error) {

	system, err := cache.ReadSystemSaveData(cache.Ctx, uuid)

	if errors.Is(err, redis.Nil) {
		if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
			system, err = r.GetSystemSaveFromS3(ctx, uuid)
		} else { // use database
			system, err = r.next.ReadSystemSaveData(ctx, uuid)
			cache.StoreSystemSaveData(ctx, uuid, system)
		}

		if err != nil {
			return system, err
		}
	}

	return system, err
}

func (r *savedataRepository) StoreSystemSaveData(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	return cache.StoreSystemSaveData(ctx, uuid, data)
}

func (r *savedataRepository) StoreSystemSaveDataS3(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	return r.next.StoreSystemSaveDataS3(ctx, uuid, data)
}

func (r *savedataRepository) DeleteSystemSaveData(ctx context.Context, uuid []byte) error {
	return r.next.DeleteSystemSaveData(ctx, uuid)
}

func (r *savedataRepository) ReadSessionSaveData(ctx context.Context, uuid []byte, slot int) (defs.SessionSaveData, error) {

	session, err := cache.ReadSessionSaveData(ctx, uuid, slot)

	if errors.Is(err, redis.Nil) {
		session, err = db.ReadSessionSaveData(uuid, slot)
	}

	return session, err
}

func (r *savedataRepository) GetLatestSessionSaveDataSlot(ctx context.Context, uuid []byte) (int, error) {
	return r.next.GetLatestSessionSaveDataSlot(ctx, uuid)
}

func (r *savedataRepository) StoreSessionSaveData(ctx context.Context, uuid []byte, data defs.SessionSaveData, slot int) error {
	return cache.StoreSessionSaveData(ctx, uuid, data, slot)
	//return db.StoreSessionSaveData(uuid, data, slot)
}

func (r *savedataRepository) DeleteSessionSaveData(ctx context.Context, uuid []byte, slot int) error {
	cache.DeleteSessionSaveData(ctx, uuid, slot)
	return r.next.DeleteSessionSaveData(ctx, uuid, slot)
}

func (r *savedataRepository) StoreSessionSaveDataBulk(ctx context.Context, uuids [][]byte, sessionsDataMapList []defs.SessionSaveData, slot int) error {
	return r.next.StoreSessionSaveDataBulk(ctx, uuids, sessionsDataMapList, slot)
}

func (r *savedataRepository) StoreSystemSaveDataBulk(ctx context.Context, uuids [][]byte, systemsDataList []defs.SystemSaveData) error {
	return r.next.StoreSystemSaveDataBulk(ctx, uuids, systemsDataList)
}

func (r *savedataRepository) RetrievePlaytime(ctx context.Context, uuid []byte) (int, error) {
	time, err := cache.RetrievePlaytime(ctx, uuid)
	if err != nil {
		time, err = db.RetrievePlaytime(uuid)
	}
	return time, err
}

func (r *savedataRepository) GetSystemSaveFromS3(ctx context.Context, uuid []byte) (defs.SystemSaveData, error) {
	return r.next.GetSystemSaveFromS3(ctx, uuid)
}
