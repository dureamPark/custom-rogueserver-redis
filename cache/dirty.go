package cache

import (
	"github.com/pagefaultgames/rogueserver/util/logger"
)

// DirtyKeysSet is the name of the Redis set that holds keys pending a write-back.
const DirtyKeysSet = "dirty_keys"

// MarkAsDirty flags one or more keys as needing a write-back to the database.
// It adds the provided keys to the dirty set in Redis.
func MarkAsDirty(keys ...string) error {

	logger.Info("MarkAsDirty")
	if len(keys) == 0 {
		return nil
	}

	// SADD can take multiple members at once, which is efficient.
	// We need to convert []string to []interface{} for the command.
	members := make([]interface{}, len(keys))
	for i, k := range keys {
		members[i] = k
	}

	logger.Info("Marking %d keys as dirty: %v", len(members), members)

	return Rdb.SAdd(Ctx, DirtyKeysSet, members...).Err()
}
