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

package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pagefaultgames/rogueserver/repository"
)

// /account/logout - log out of account
func Logout(context context.Context, token []byte) error {
	// err := db.RemoveSessionFromToken(token)
	err := repository.Repos.Account.RemoveSessionFromToken(context, token)
	// TODO. 남아있는 데이터를 로그아웃할 때, 전부 DB에 저장할건지, Cache 정책에 따라 저장할 것인지
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("token not found")
		}

		return fmt.Errorf("failed to remove account session")
	}

	return nil
}
