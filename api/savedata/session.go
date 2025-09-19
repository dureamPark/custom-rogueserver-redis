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
)

func GetSession(context context.Context, uuid []byte, slot int) (defs.SessionSaveData, error) {
	session, err := repository.Repos.Savedata.ReadSessionSaveData(context, uuid, slot)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrSaveNotExist
		}

		return session, err
	}

	return session, nil
}

func UpdateSession(context context.Context, uuid []byte, slot int, data defs.SessionSaveData) error {
	//err := db.StoreSessionSaveData(uuid, data, slot)
	err := repository.Repos.Savedata.StoreSessionSaveData(context, uuid, data, slot)
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
