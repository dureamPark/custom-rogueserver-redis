package worker

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"

	"github.com/pagefaultgames/rogueserver/db"
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

		base64String := strings.TrimPrefix(key, "session:")
		uuid, err := base64.StdEncoding.DecodeString(base64String)

		// TODO.언마샬링할 때, Account에 있는 Time.time 자료형으로 인해 에러 발생
		// 일단 필요한 데이터만 받아서 사용하는 식으로 구현해두고, 나중에 캐시바꾸고 여기도 바꾸기
		// dataTypeArr := [3]string{"accountStats", "systemSaveData", "sessionSaveData"}

		defaultPath := "$."
		systemPath := defaultPath + "systemSaveData"
		systemDataJSON, err := w.redisClient.JSONGet(ctx, key, systemPath).Result()
		if err != nil {
			logger.Error("Error JSONGET system savedata for key %s: %v", key, err)
			continue
		}

		var systemData defs.SystemSaveData
		var systemDataArr []defs.SystemSaveData
		err = json.Unmarshal([]byte(systemDataJSON), &systemDataArr)
		if err != nil {
			logger.Error("Error unmarshaling system savedata for key %s: %v", key, err)
			continue
		}

		if len(systemDataArr) > 0 {
			systemData = systemDataArr[0]
		} else {
			continue
		}

		// AccountStat1s - systemData.VoucherCounts이 NULL이라 뜨네..
		// err = db.UpdateAccountStats(uuid, systemData.GameStats, systemData.VoucherCounts)
		// if err != nil {
		// 	logger.Error("WriteBack - UpdateAccountStats Error : %s", err)
		// 	continue
		// }

		// SystemData
		err = db.StoreSystemSaveData(uuid, systemData)
		if err != nil {
			logger.Error("WriteBack - StoreSystemSaveData Error : %s", err)
			continue
		}

		// Session
		sessionPath := defaultPath + "sessionSaveData"
		sessionDataJSON, err := w.redisClient.JSONGet(ctx, key, sessionPath).Result()
		if err != nil {
			logger.Error("Error JSONGET session savedata for key %s: %v", key, err)
			continue
		}

		var sessionDataMap map[string]defs.SessionSaveData
		var sessionDataArr []map[string]defs.SessionSaveData
		err = json.Unmarshal([]byte(sessionDataJSON), &sessionDataArr)
		if err != nil {
			logger.Error("Error unmarshaling session savedata for key %s: %v", key, err)
			continue
		}

		if len(sessionDataArr) > 0 {
			sessionDataMap = sessionDataArr[0]
		} else {
			continue
		}

		// sessiondata
		index := 0
		for _, sessionData := range sessionDataMap {
			err = db.StoreSessionSaveData(uuid, sessionData, index)
			if err != nil {
				logger.Error("WriteBack - StoreSessionSaveData Error : %s", err)
			}
			index++
		}

		logger.Info("Successfully wrote key %s to database.", key)
	}
}
