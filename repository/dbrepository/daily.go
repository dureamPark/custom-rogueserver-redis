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
	"math"

	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/repository"
)

// dailyDBRepository는 DailyRepository 인터페이스의 데이터베이스 구현체입니다.
type dailyDBRepository struct {
	db *sql.DB
}

// NewDailyDBRepository는 데이터베이스와 통신하는 새로운 daily repository를 생성합니다.
func NewDailyDBRepository(db *sql.DB) repository.DailyRepository {
	return &dailyDBRepository{db: db}
}

// TryAddDailyRun은 오늘의 일일 실행을 데이터베이스에 추가하려고 시도합니다.
// MariaDB 10.5+는 ON DUPLICATE KEY UPDATE와 함께 RETURNING을 지원합니다.
func (r *dailyDBRepository) TryAddDailyRun(ctx context.Context, seed string) (string, error) {
	var actualSeed string
	query := `
		INSERT INTO dailyRuns (seed, date) VALUES (?, UTC_DATE())
		ON DUPLICATE KEY UPDATE date = date
		RETURNING seed`
	err := r.db.QueryRowContext(ctx, query, seed).Scan(&actualSeed)
	if err != nil {
		return "", err
	}
	return actualSeed, nil
}

// GetDailyRunSeed는 오늘 날짜의 일일 실행 시드를 가져옵니다.
func (r *dailyDBRepository) GetDailyRunSeed(ctx context.Context) (string, error) {
	var seed string
	err := r.db.QueryRowContext(ctx, "SELECT seed FROM dailyRuns WHERE date = UTC_DATE()").Scan(&seed)
	if err != nil {
		return "", err
	}

	return seed, nil
}

// AddOrUpdateAccountDailyRun은 계정의 일일 실행 기록을 추가하거나 업데이트합니다.
func (r *dailyDBRepository) AddOrUpdateAccountDailyRun(ctx context.Context, uuid []byte, score int, wave int) error {
	query := `
		INSERT INTO accountDailyRuns (uuid, date, score, wave, timestamp) 
		VALUES (?, UTC_DATE(), ?, ?, UTC_TIMESTAMP()) 
		ON DUPLICATE KEY UPDATE 
			score = GREATEST(score, ?), 
			wave = GREATEST(wave, ?), 
			timestamp = IF(score < ?, UTC_TIMESTAMP(), timestamp)`
	_, err := r.db.ExecContext(ctx, query, uuid, score, wave, score, wave, score)
	return err
}

// FetchRankings는 지정된 카테고리와 페이지에 대한 일일 순위를 가져옵니다.
func (r *dailyDBRepository) FetchRankings(ctx context.Context, category int, page int) ([]defs.DailyRanking, error) {
	var rankings []defs.DailyRanking
	offset := (page - 1) * 10

	var query string
	switch category {
	case 0:
		query = `
			SELECT 
				RANK() OVER (ORDER BY adr.score DESC, adr.timestamp), 
				a.username, 
				adr.score, 
				adr.wave 
			FROM accountDailyRuns adr 
			JOIN dailyRuns dr ON dr.date = adr.date 
			JOIN accounts a ON adr.uuid = a.uuid 
			WHERE dr.date = UTC_DATE() AND a.banned = 0 
			LIMIT 10 OFFSET ?`
	case 1:
		query = `
			SELECT 
				RANK() OVER (ORDER BY SUM(adr.score) DESC, adr.timestamp), 
				a.username, 
				SUM(adr.score), 
				0 
			FROM accountDailyRuns adr 
			JOIN dailyRuns dr ON dr.date = adr.date 
			JOIN accounts a ON adr.uuid = a.uuid 
			WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY) AND a.banned = 0 
			GROUP BY a.username 
			ORDER BY 1 
			LIMIT 10 OFFSET ?`
	}

	results, err := r.db.QueryContext(ctx, query, offset)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	for results.Next() {
		var ranking defs.DailyRanking
		err = results.Scan(&ranking.Rank, &ranking.Username, &ranking.Score, &ranking.Wave)
		if err != nil {
			return nil, err
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

// FetchRankingPageCount는 지정된 카테고리의 총 순위 페이지 수를 가져옵니다.
func (r *dailyDBRepository) FetchRankingPageCount(ctx context.Context, category int) (int, error) {
	var query string
	switch category {
	case 0:
		query = `
			SELECT COUNT(a.username) 
			FROM accountDailyRuns adr 
			JOIN dailyRuns dr ON dr.date = adr.date 
			JOIN accounts a ON adr.uuid = a.uuid 
			WHERE dr.date = UTC_DATE()`
	case 1:
		query = `
			SELECT COUNT(DISTINCT a.username) 
			FROM accountDailyRuns adr 
			JOIN dailyRuns dr ON dr.date = adr.date 
			JOIN accounts a ON adr.uuid = a.uuid 
			WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY)`
	}

	var recordCount int
	err := r.db.QueryRowContext(ctx, query).Scan(&recordCount)
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(float64(recordCount) / 10)), nil
}
