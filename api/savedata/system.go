/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package savedata

import (
	"fmt"
	"os"
	"log"
	"encoding/base64"
	"errors"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/redis/go-redis/v9"
)

func GetSystem(uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData
	var err error

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)

	system, err = cache.ReadSystemSaveData(uuid)

	if errors.Is(err, redis.Nil) {
		// 캐시에 저장된 세션 정보가 없으면
		log.Printf("시스템 정보가 캐시에 없습니다.(key : %s) : %s", encodedUUID, err)

		if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
			system, err = db.GetSystemSaveFromS3(uuid)
		} else { // use database
			//log.Println("use database GetSystem");
			system, err = db.ReadSystemSaveData(uuid)
		}
		log.Printf("시스템 정보를 DB에서 찾습니다.")

		if err != nil {
			return system, err
		}
	}

	return system, nil
}

func UpdateSystem(uuid []byte, data defs.SystemSaveData) error {
	if data.TrainerId == 0 && data.SecretId == 0 {
		return fmt.Errorf("invalid system data")
	}

	err := db.UpdateAccountStats(uuid, data.GameStats, data.VoucherCounts)
	if err != nil {
		return fmt.Errorf("failed to update account stats: %s", err)
	}

	if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
		err = db.StoreSystemSaveDataS3(uuid, data)
	} else {
		err = db.StoreSystemSaveData(uuid, data)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeleteSystem(uuid []byte) error {
	err := db.DeleteSystemSaveData(uuid)
	if err != nil {
		return err
	}

	return nil
}
