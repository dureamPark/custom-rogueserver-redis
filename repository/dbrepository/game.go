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

package dbrepository

import (
	"context"
	"database/sql"

	"github.com/pagefaultgames/rogueserver/repository"
)

// gameDBRepository는 GameRepository 인터페이스의 데이터베이스 구현체입니다.
type gameDBRepository struct {
	db *sql.DB
}

// NewGameDBRepository는 데이터베이스와 통신하는 새로운 game repository를 생성합니다.
func NewGameDBRepository(db *sql.DB) repository.GameRepository {
	return &gameDBRepository{db: db}
}

func (r *gameDBRepository) FetchPlayerCount(ctx context.Context) (int, error) {
	var playerCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM accounts WHERE lastActivity > DATE_SUB(UTC_TIMESTAMP(), INTERVAL 5 MINUTE)").Scan(&playerCount)
	if err != nil {
		return 0, err
	}

	return playerCount, nil
}

func (r *gameDBRepository) FetchBattleCount(ctx context.Context) (int, error) {
	var battleCount int
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(s.battles), 0) FROM accountStats s JOIN accounts a ON a.uuid = s.uuid WHERE a.banned = 0").Scan(&battleCount)
	if err != nil {
		return 0, err
	}

	return battleCount, nil
}

func (r *gameDBRepository) FetchClassicSessionCount(ctx context.Context) (int, error) {
	var classicSessionCount int
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(s.classicSessionsPlayed), 0) FROM accountStats s JOIN accounts a ON a.uuid = s.uuid WHERE a.banned = 0").Scan(&classicSessionCount)
	if err != nil {
		return 0, err
	}

	return classicSessionCount, nil
}
