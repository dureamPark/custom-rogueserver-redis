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
	"context"
	"database/sql"
	"errors"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/repository"
	"github.com/pagefaultgames/rogueserver/util/logger"
)

func GetSession(context context.Context, uuid []byte, slot int, level string) (defs.SessionSaveData, error) {
	var session defs.SessionSaveData
	var err error
	if level == "Cache" {
		session, err = repository.Repos.Savedata.ReadSessionSaveData(context, uuid, slot)
	} else {
		session, err = repository.ReposDB.Savedata.ReadSessionSaveData(context, uuid, slot)
	}
	//session, err = repository.Repos.Savedata.ReadSessionSaveData(context, uuid, slot);
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Error("Level : %s", level)
			err = ErrSaveNotExist
		}

		return session, err
	}

	return session, nil
}

func UpdateSession(context context.Context, uuid []byte, slot int, data defs.SessionSaveData, level string) error {
	//err := db.StoreSessionSaveData(uuid, data, slot)
	var err error
	if level == "Cache" {
		err = repository.Repos.Savedata.StoreSessionSaveData(context, uuid, data, slot)
	} else {
		err = repository.ReposDB.Savedata.StoreSessionSaveData(context, uuid, data, slot)
	}

	//err := repository.Repos.Savedata.StoreSessionSaveData(context, uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSession(context context.Context, uuid []byte, slot int) error {
	//err := db.DeleteSessionSaveData(uuid, slot)
	err := repository.Repos.Savedata.DeleteSessionSaveData(context, uuid, slot)
	if err != nil {
		return err
	}

	return nil
}
