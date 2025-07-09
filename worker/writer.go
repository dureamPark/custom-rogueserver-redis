package worker

import (
	"context"
	"database/sql"
	//"encoding/json"
	"log"
	//"strings"
	"time"

	"github.com/redis/go-redis/v9"
	//"github.com/pagefaultgames/rogueserver/db"
	//"github.com/pagefaultgames/rogueserver/defs"
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
	log.Println("Starting write-back worker...")
	
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
	log.Println("Starting write-back worker...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping write-back worker...")
			return
		case <-ticker.C:
			w.flushDirtyData(ctx)
		}
	}
}

func (w *WriteBackWorker) flushDirtyData(ctx context.Context) {
	// keys, err := w.redisClient.SPopN(ctx, dirtyKeysSet, batchSize).Result()
	// if err != nil {
	// 	if err != redis.Nil {
	// 		log.Printf("Error popping dirty keys from Redis: %v", err)
	// 	}
	// 	return
	// }

	// if len(keys) == 0 {
	// 	return
	// }

	// log.Printf("Processing %d dirty keys...", len(keys))

	// for _, key := range keys {
	// 	data, err := w.redisClient.Get(ctx, key).Bytes()
	// 	// if err != nil {
	// 	// 	log.Printf("Error getting data for key %s: %v", key, err)
	// 	// 	continue
	// 	// }

	// 	// // Parse the key to determine data type
	// 	// parts := strings.SplitN(key, ":", 2)
	// 	// if len(parts) != 2 {
	// 	// 	log.Printf("Invalid key format, skipping: %s", key)
	// 	// 	continue
	// 	// }
	// 	// dataType := parts[0]

	// 	// var writeErr error
	// 	// // Route to the correct handler based on data type
	// 	// switch dataType {
	// 	// case "savedata":
	// 	// 	var saveData defs.SaveData
	// 	// 	if err := json.Unmarshal(data, &saveData); err != nil {
	// 	// 		log.Printf("Error unmarshaling savedata for key %s: %v", key, err)
	// 	// 		continue
	// 	// 	}
	// 	// 	writeErr = db.UpdateSaveData(ctx, w.db, &saveData)

	// 	// case "account":
	// 	// 	var accountData defs.Account
	// 	// 	if err := json.Unmarshal(data, &accountData); err != nil {
	// 	// 		log.Printf("Error unmarshaling account data for key %s: %v", key, err)
	// 	// 		continue
	// 	// 	}
	// 	// 	// Assuming you have a function like this in your db package
	// 	// 	writeErr = db.UpdateAccount(ctx, w.db, &accountData)
		
	// 	// // Add other cases for other data types here
	// 	// // case "guild":
	// 	// // ...

	// 	// default:
	// 	// 	log.Printf("Unknown data type '%s' for key: %s", dataType, key)
	// 	// 	continue
	// 	// }

	// 	// Common error handling for DB write operations
	// 	if writeErr != nil {
	// 		log.Printf("Error writing data for key %s to DB: %v", key, writeErr)
	// 		// If DB write fails, add the key back to the dirty set to retry.
	// 		// if err := w.redisClient.SAdd(ctx, dirtyKeysSet, key).Err(); err != nil {
	// 		// 	log.Printf("CRITICAL: Failed to add key %s back to dirty set after DB error: %v", key, err)
	// 		// }
	// 		// continue
	// 	}
	// 	log.Printf("Successfully wrote key %s to database.", key)
	// }
}
