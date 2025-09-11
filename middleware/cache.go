package middleware

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"
	"github.com/redis/go-redis/v9"
)

func GetSessionCache(uuid []byte, slot int) (defs.SessionSaveData, error) {

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)

	session, err := cache.ReadSessionSaveData(uuid, slot)

	if errors.Is(err, redis.Nil) {
		// 캐시에 저장된 세션 정보가 없으면
		logger.Info("세션 정보가 캐시에 없습니다.(key : %s, slot : %d) : %s", encodedUUID, slot, err)
		session, err = db.ReadSessionSaveData(uuid, slot)
		logger.Info("세션 정보를 DB에서 찾습니다.")

		if err == nil {
			// DB에서 세션 값을 가져왔을 때만
			UpdateSessionCache(uuid, slot, session)
		}
	}

	if err != nil {
		logger.Error("Fail to Get Session (key : %s) : %s | session : %s", encodedUUID, err, session)
		return session, err
	}

	return session, nil
}

func UpdateCache(uuid []byte, slot int, save any) error {
	err := cache.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	switch save := save.(type) {
	case defs.SystemSaveData: // System
		if save.TrainerId == 0 && save.SecretId == 0 {
			return fmt.Errorf("invalid system data")
		}

		err = cache.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		//err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		if err != nil {
			return fmt.Errorf("failed to update account stats: %s", err)
		}
		//return db.StoreSystemSaveData(uuid, save)
		return cache.StoreSystemSaveData(uuid, save)

	case defs.SessionSaveData: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return fmt.Errorf("slot id %d out of range", slot)
		}
		//return db.StoreSessionSaveData(uuid, save, slot)
		return cache.StoreSessionSaveData(uuid, save, slot)

	default:
		return fmt.Errorf("invalid data type")
	}
}

func UpdateSessionCache(uuid []byte, slot int, data defs.SessionSaveData) error {
	err := cache.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}
