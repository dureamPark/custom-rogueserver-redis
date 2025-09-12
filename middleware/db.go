package middleware

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"
)

func GetSessionDB(uuid []byte, slot int) (defs.SessionSaveData, error) {

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)

	session, err := db.ReadSessionSaveData(uuid, slot)

	if err == nil {
		// DB에서 세션 값을 가져왔을 때만
		UpdateSessionDB(uuid, slot, session)
	}

	if err != nil {
		logger.Error("Fail to Get Session (key : %s) : %s | session : %s", encodedUUID, err, session)
		return session, err
	}

	return session, nil
}

func UpdateDB(uuid []byte, slot int, save any) error {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	switch save := save.(type) {
	case defs.SystemSaveData: // System
		if save.TrainerId == 0 && save.SecretId == 0 {
			return fmt.Errorf("invalid system data")
		}

		err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		//err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		if err != nil {
			return fmt.Errorf("failed to update account stats: %s", err)
		}
		return db.StoreSystemSaveData(uuid, save)

	case defs.SessionSaveData: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return fmt.Errorf("slot id %d out of range", slot)
		}
		return db.StoreSessionSaveData(uuid, save, slot)

	default:
		return fmt.Errorf("invalid data type")
	}
}

func UpdateSessionDB(uuid []byte, slot int, data defs.SessionSaveData) error {
	err := db.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

func GetProfileScore() (float64, error) {
	return 0, nil
}
