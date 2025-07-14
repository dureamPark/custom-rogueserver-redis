package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/pagefaultgames/rogueserver/util/logger"
	"time"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

const (
	dirtyKeysSet = "dirty_keys"
	batchSize    = 100 // Number of keys to process in one batch
)

// WriteBackWorker periodically flushes dirty data from cache to the persistent database.
type WriteBackWorker struct {
	redisClient *redis.Client
	db          *sql.DB
}

// NewWriteBackWorker creates a new worker instance.
func NewWriteBackWorker(redisClient *redis.Client, db *sql.DB) *WriteBackWorker {
	return &WriteBackWorker{
		redisClient: redisClient,
		db:          db,
	}
}

func StartWriteBackWorker(db *sql.DB, redisClient *redis.Client) {
	logger.Info("Starting write-back worker...")

	worker := NewWriteBackWorker(redisClient, db)

	// Start the worker in a separate goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go worker.Run(ctx)

	// Ensure we stop the worker gracefully on shutdown
	go func() {
		// Wait for a signal to stop (e.g., SIGINT, SIGTERM)
		// This is just an example; you should implement proper signal handling.
		<-time.After(24 * time.Hour) // Replace with actual signal handling
		cancel()
	}()

}

// Run starts the worker's main loop in a goroutine.
func (w *WriteBackWorker) Run(ctx context.Context) {
	logger.Info("Starting write-back worker...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping write-back worker...")
			return
		case <-ticker.C:
			w.flushDirtyData(ctx)
		}
	}
}

func (w *WriteBackWorker) flushDirtyData(ctx context.Context) {
	logger.Info("flush Dirty Data")

	// dirty key 확인하고 db에 업데이트해주기

	// dirtyKeysSet에 있는 key들 batchSize만큼 가져오기
	keys, err := w.redisClient.SPopN(ctx, dirtyKeysSet, batchSize).Result()

	if err != nil {
		// key가 없거나 데이터가 없는 경우
		if err != redis.Nil {
			logger.Error("Error popping dirty keys from Redis: %v", err)
		}
		return
	}

	// 변경된 데이터가 없는 경우 = dirtyKeys가 없는 경우
	if len(keys) == 0 {
		return
	}

	logger.Info("Processing %d dirty keys...", len(keys))

	// dirtyKeys에 해당하는 값들 가져오기
	for _, key := range keys {
		keyBytes := []byte(key)
		logger.Info(key)
		// uuid 가져오기
		parts := bytes.SplitN(keyBytes, []byte(":"), 2)
		var uuid []byte
		if len(parts) == 2 {
			uuid = parts[1]
			logger.Info("추출된 UUID : %s", uuid)
		} else {
			logger.Error("잘못된 키 형식 입니다.")
			continue
		}

		// 현재는 하나의 세션과 id로 묶여있으니까 그냥 JSON으로 가져오면 됨.
		// bytes() -> unmarshal -> struct
		jsonStr, err := w.redisClient.JSONGet(ctx, key, "$").Result()
		if err != nil {
			logger.Error("Error getting JSON data for key %s: %v", key, err)
			continue
		}

		data := []byte(jsonStr)

		var dirtyData defs.UserCacheData
		if err := json.Unmarshal(data, &dirtyData); err != nil {
			logger.Error("Error unmarshaling savedata for key %s: %v", key, err)
			continue
		}

		// AccountStats
		err = db.UpdateAccountStats(uuid, dirtyData.SystemSaveData.GameStats, dirtyData.SystemSaveData.VoucherCounts)
		if err != nil {
			logger.Error("WriteBack - UpdateAccountStats Error : %s", err)
		}

		err = db.StoreSystemSaveData(uuid, *dirtyData.SystemSaveData)
		if err != nil {
			logger.Error("WriteBack - StoreSystemSaveData Error : %s", err)
		}

		// sessiondata
		index := 0
		for _, sessionData := range dirtyData.SessionSaveData {
			err = db.StoreSessionSaveData(uuid, sessionData, index)
			if err != nil {
				logger.Error("WriteBack - StoreSessionSaveData Error : %s", err)
			}
			index++
		}

		logger.Info("Successfully wrote key %s to database.", key)
	}
}
