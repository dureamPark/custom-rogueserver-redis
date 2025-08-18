package worker

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"

	"github.com/redis/go-redis/v9"
)

const (
	dirtyKeysSet = "dirty_keys"
	batchSize    = 10000 // Number of keys to process in one batch
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

	keys, err := w.redisClient.SPopN(ctx, dirtyKeysSet, batchSize).Result()
	if err != nil && err != redis.Nil {
		logger.Error("Error popping dirty keys from Redis: %v", err)
		return
	}

	if len(keys) == 0 {
		return
	}

	logger.Info("Processing %d dirty keys...", len(keys))

	// 1. 모든 키에 대해 systemSaveData를 한 번에 가져오기
	systemDataJSONs, err := w.redisClient.JSONMGet(ctx, ".systemSaveData", keys...).Result()
	if err != nil {
		logger.Error("Error JSONMGet system savedata: %v", err)
		return
	}

	// 2. 모든 키에 대해 sessionSaveData를 한 번에 가져오기
	sessionDataJSONs, err := w.redisClient.JSONMGet(ctx, ".sessionSaveData", keys...).Result()
	if err != nil {
		logger.Error("Error JSONMGet session savedata: %v", err)
		return
	}

	// Bulk 업데이트를 위해 데이터를 담을 슬라이스 정의
	var systemDataList []defs.SystemSaveData
	var sessionDataMapList []map[string]defs.SessionSaveData
	var uuidList [][]byte

	// 3. 가져온 데이터를 순회하며 디코딩 및 슬라이스에 추가
	for i, key := range keys {
		// UUID 디코딩
		base64String := strings.TrimPrefix(key, "session:")
		uuid, err := base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			logger.Error("Error decoding UUID for key %s: %v", key, err)
			continue
		}
		uuidList = append(uuidList, uuid)

		// systemSaveData 디코딩
		var systemData defs.SystemSaveData
		if systemDataJSONs[i] != nil {
			err = json.Unmarshal([]byte(systemDataJSONs[i].(string)), &systemData)
			if err != nil {
				logger.Error("Error unmarshaling system savedata for key %s: %s | %v", key, systemDataJSONs[i], err)
				continue
			}
			systemDataList = append(systemDataList, systemData)
		} else {
			// 데이터가 없는 경우 로깅 또는 건너뛰기
			logger.Warn("System data not found for key %s", key)
			continue
		}

		// sessionSaveData 디코딩
		var sessionDataMap map[string]defs.SessionSaveData
		if sessionDataJSONs[i] != nil {
			err = json.Unmarshal([]byte(sessionDataJSONs[i].(string)), &sessionDataMap)
			if err != nil {
				logger.Error("Error unmarshaling session savedata for key %s: %s | %v", key, sessionDataJSONs[i], err)
				continue
			}
			sessionDataMapList = append(sessionDataMapList, sessionDataMap)
		} else {
			// 데이터가 없는 경우 로깅 또는 건너뛰기
			logger.Warn("Session data not found for key %s", key)
			continue
		}
	}
	logger.Info("len(systemDataList) : %d", len(systemDataList))

	// 4. 모든 데이터가 준비되면 한 번의 DB 호출로 업데이트
	if len(systemDataList) > 0 {
		err = db.StoreSystemSaveDataBulk(ctx, uuidList, systemDataList)
		if err != nil {
			logger.Error("WriteBack - Bulk StoreSystemSaveData Error : %s", err)
		}
	}

	if len(sessionDataMapList) > 0 {
		err = db.StoreSessionSaveDataBulk(ctx, uuidList, sessionDataMapList, 0)
		if err != nil {
			logger.Error("WriteBack - Bulk StoreSessionSaveDataBulk Error : %s", err)
		}
	}

	logger.Info("Successfully wrote %d keys to database.", len(keys))
}
