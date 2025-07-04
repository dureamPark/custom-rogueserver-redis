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
	"encoding/base64"
	"errors"
	"log"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/redis/go-redis/v9"
)

func GetSession(uuid []byte, slot int) (defs.SessionSaveData, error) {
	log.Println("GetSession")

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)

	session, err := cache.ReadSessionSaveData(uuid, slot)

	if errors.Is(err, redis.Nil) {
		// 캐시에 저장된 세션 정보가 없으면
		log.Printf("세션 정보가 캐시에 없습니다.(key : %s) : %s", encodedUUID, err)
		session, err = db.ReadSessionSaveData(uuid, slot)
		log.Printf("세션 정보를 DB에서 찾습니다.")

		if err == nil {
			// DB에서 세션 값을 가져왔을 때만
			UpdateSession(uuid, slot, session)
		}
	}

	if err != nil {
		log.Printf("Fail to Get Session (key : %s) : %s", encodedUUID, err)
		return session, err
	}

	return session, nil
}

func UpdateSession(uuid []byte, slot int, data defs.SessionSaveData) error {
	//err := db.StoreSessionSaveData(uuid, data, slot)
	err := cache.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSession(uuid []byte, slot int) error {
	//err := db.DeleteSessionSaveData(uuid, slot)
	err := cache.DeleteSessionSaveData(uuid, slot)
	if err != nil {
		return err
	}

	return nil
}
