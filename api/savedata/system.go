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
	"fmt"
	"os"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/repository"
)

func GetSystem(context context.Context, uuid []byte) (defs.SystemSaveData, error) {

	system, err := repository.Repos.Savedata.ReadSystemSaveData(context, uuid)

	return system, err
}

func UpdateSystem(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	if data.TrainerId == 0 && data.SecretId == 0 {
		return fmt.Errorf("invalid system data")
	}

	err := repository.Repos.Account.UpdateAccountStats(ctx, uuid, data.GameStats, data.VoucherCounts)
	if err != nil {
		return fmt.Errorf("failed to update account stats: %s", err)
	}

	if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
		err = repository.Repos.Savedata.StoreSystemSaveDataS3(ctx, uuid, data)
	} else {
		err = repository.Repos.Savedata.StoreSystemSaveData(ctx, uuid, data)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeleteSystem(ctx context.Context, uuid []byte) error {
	err := repository.Repos.Savedata.DeleteSystemSaveData(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
