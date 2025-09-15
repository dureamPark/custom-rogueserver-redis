package repository

import (
	"context"
	"database/sql"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// savedataRepository 구조체 선언 (DB 핸들 포함)
type savedataRepository struct {
	db *sql.DB
}

// 생성자 함수: DB 핸들 주입
func NewSavedataRepository(db *sql.DB) *savedataRepository {
	return &savedataRepository{db: db}
}

// 아래부터 인터페이스 구현 (context는 받지만 db 함수에는 넘기지 않음)

func (r *savedataRepository) TryAddSeedCompletion(ctx context.Context, uuid []byte, seed string, mode int) (bool, error) {
	return db.TryAddSeedCompletion(uuid, seed, mode)
}

func (r *savedataRepository) ReadSeedCompleted(ctx context.Context, uuid []byte, seed string) (bool, error) {
	return db.ReadSeedCompleted(uuid, seed)
}

func (r *savedataRepository) ReadSystemSaveData(ctx context.Context, uuid []byte) (defs.SystemSaveData, error) {
	return db.ReadSystemSaveData(uuid)
}

func (r *savedataRepository) StoreSystemSaveData(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	return db.StoreSystemSaveData(uuid, data)
}

func (r *savedataRepository) StoreSystemSaveDataS3(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	return db.StoreSystemSaveDataS3(uuid, data)
}

func (r *savedataRepository) DeleteSystemSaveData(ctx context.Context, uuid []byte) error {
	return db.DeleteSystemSaveData(uuid)
}

func (r *savedataRepository) ReadSessionSaveData(ctx context.Context, uuid []byte, slot int) (defs.SessionSaveData, error) {
	return db.ReadSessionSaveData(uuid, slot)
}

func (r *savedataRepository) GetLatestSessionSaveDataSlot(ctx context.Context, uuid []byte) (int, error) {
	return db.GetLatestSessionSaveDataSlot(uuid)
}

func (r *savedataRepository) StoreSessionSaveData(ctx context.Context, uuid []byte, data defs.SessionSaveData, slot int) error {
	return db.StoreSessionSaveData(uuid, data, slot)
}

func (r *savedataRepository) DeleteSessionSaveData(ctx context.Context, uuid []byte, slot int) error {
	return db.DeleteSessionSaveData(uuid, slot)
}

func (r *savedataRepository) StoreSessionSaveDataBulk(ctx context.Context, uuids [][]byte, sessionsDataMapList []defs.SessionSaveData, slot int) error {
	return db.StoreSessionSaveDataBulk(ctx, uuids, sessionsDataMapList, slot)
}

func (r *savedataRepository) StoreSystemSaveDataBulk(ctx context.Context, uuids [][]byte, systemsDataList []defs.SystemSaveData) error {
	return db.StoreSystemSaveDataBulk(ctx, uuids, systemsDataList)
}

func (r *savedataRepository) RetrievePlaytime(ctx context.Context, uuid []byte) (int, error) {
	return db.RetrievePlaytime(uuid)
}

func (r *savedataRepository) GetSystemSaveFromS3(ctx context.Context, uuid []byte) (defs.SystemSaveData, error) {
	return db.GetSystemSaveFromS3(uuid)
}
