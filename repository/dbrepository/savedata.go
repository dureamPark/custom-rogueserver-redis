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
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/klauspost/compress/zstd"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/repository"
)

// savedataDBRepository는 SaveDataRepository 인터페이스의 데이터베이스 구현체입니다.
type savedataDBRepository struct {
	db *sql.DB
	// TODO: accountRepository를 주입하여 FetchUsernameFromUUID를 호출해야 합니다.
	accountRepo repository.AccountRepository
}

// NewSaveDataDBRepository는 데이터베이스와 통신하는 새로운 saveData repository를 생성합니다.
func NewSavedataDBRepository(db *sql.DB, account repository.AccountRepository) repository.SavedataRepository {
	return &savedataDBRepository{db: db, accountRepo: account}
}

func (r *savedataDBRepository) TryAddSeedCompletion(ctx context.Context, uuid []byte, seed string, mode int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dailyRunCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	} else if count > 0 {
		return false, nil
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO dailyRunCompletions (uuid, seed, mode, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP())", uuid, seed, mode)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *savedataDBRepository) ReadSeedCompleted(ctx context.Context, uuid []byte, seed string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dailyRunCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *savedataDBRepository) ReadSystemSaveData(ctx context.Context, uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData

	var data []byte
	err := r.db.QueryRowContext(ctx, "SELECT data FROM systemSaveData WHERE uuid = ?", uuid).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return system, nil // 저장 데이터가 없으면 빈 구조체와 nil 에러 반환
		}
		return system, err
	}

	zr, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return system, err
	}
	defer zr.Close()

	err = gob.NewDecoder(zr).Decode(&system)
	if err != nil {
		return system, err
	}

	return system, nil
}

func (r *savedataDBRepository) StoreSystemSaveData(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	buf := new(bytes.Buffer)

	zw, err := zstd.NewWriter(buf)
	if err != nil {
		return err
	}

	err = gob.NewEncoder(zw).Encode(data)
	if err != nil {
		zw.Close()
		return err
	}
	zw.Close()

	_, err = r.db.ExecContext(ctx, "REPLACE INTO systemSaveData (uuid, data, timestamp) VALUES (?, ?, UTC_TIMESTAMP())", uuid, buf.Bytes())
	return err
}

func (r *savedataDBRepository) StoreSystemSaveDataS3(ctx context.Context, uuid []byte, data defs.SystemSaveData) error {
	cfg, _ := config.LoadDefaultConfig(ctx)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL_S3"))
	})

	username, err := r.accountRepo.FetchUsernameFromUUID(ctx, uuid)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("S3_SYSTEM_BUCKET_NAME")),
		Key:    aws.String(username),
		Body:   buf,
	})
	return err
}

func (r *savedataDBRepository) DeleteSystemSaveData(ctx context.Context, uuid []byte) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM systemSaveData WHERE uuid = ?", uuid)
	return err
}

func (r *savedataDBRepository) ReadSessionSaveData(ctx context.Context, uuid []byte, slot int) (defs.SessionSaveData, error) {
	var session defs.SessionSaveData

	var data []byte
	err := r.db.QueryRowContext(ctx, "SELECT data FROM sessionSaveData WHERE uuid = ? AND slot = ?", uuid, slot).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return session, nil // 저장 데이터가 없으면 빈 구조체와 nil 에러 반환
		}
		return session, err
	}

	zr, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return session, err
	}
	defer zr.Close()

	err = gob.NewDecoder(zr).Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func (r *savedataDBRepository) GetLatestSessionSaveDataSlot(ctx context.Context, uuid []byte) (int, error) {
	var slot int
	err := r.db.QueryRowContext(ctx, "SELECT slot FROM sessionSaveData WHERE uuid = ? ORDER BY timestamp DESC, slot ASC LIMIT 1", uuid).Scan(&slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil // 세션 데이터 없음
		}
		return -1, err
	}
	return slot, nil
}

func (r *savedataDBRepository) StoreSessionSaveData(ctx context.Context, uuid []byte, data defs.SessionSaveData, slot int) error {
	buf := new(bytes.Buffer)

	zw, err := zstd.NewWriter(buf)
	if err != nil {
		return err
	}

	err = gob.NewEncoder(zw).Encode(data)
	if err != nil {
		zw.Close()
		return err
	}
	zw.Close()

	_, err = r.db.ExecContext(ctx, "REPLACE INTO sessionSaveData (uuid, slot, data, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP())", uuid, slot, buf.Bytes())
	return err
}

func (r *savedataDBRepository) DeleteSessionSaveData(ctx context.Context, uuid []byte, slot int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessionSaveData WHERE uuid = ? AND slot = ?", uuid, slot)
	return err
}

func (r *savedataDBRepository) StoreSessionSaveDataBulk(ctx context.Context, uuids [][]byte, sessionsDataMapList []defs.SessionSaveData, slot int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var args []interface{}
	var valuePlaceholders []string

	for i, sessionsData := range sessionsDataMapList {
		uuid := uuids[i]
		var buf bytes.Buffer
		zw, err := zstd.NewWriter(&buf)
		if err != nil {
			return err
		}
		err = gob.NewEncoder(zw).Encode(sessionsData)
		if err != nil {
			zw.Close()
			return err
		}
		zw.Close()
		valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, UTC_TIMESTAMP())")
		args = append(args, uuid, slot, buf.Bytes())
	}

	if len(args) == 0 {
		return tx.Commit()
	}

	query := "REPLACE INTO sessionSaveData (uuid, slot, data, timestamp) VALUES " + strings.Join(valuePlaceholders, ", ")
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *savedataDBRepository) StoreSystemSaveDataBulk(ctx context.Context, uuids [][]byte, systemsDataList []defs.SystemSaveData) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var args []interface{}
	var valuePlaceholders []string

	for i, data := range systemsDataList {
		uuid := uuids[i]
		var buf bytes.Buffer
		zw, err := zstd.NewWriter(&buf)
		if err != nil {
			return err
		}
		err = gob.NewEncoder(zw).Encode(data)
		if err != nil {
			zw.Close()
			return err
		}
		zw.Close()
		valuePlaceholders = append(valuePlaceholders, "(?, ?, UTC_TIMESTAMP())")
		args = append(args, uuid, buf.Bytes())
	}

	if len(args) == 0 {
		return tx.Commit()
	}

	query := "REPLACE INTO systemSaveData (uuid, data, timestamp) VALUES " + strings.Join(valuePlaceholders, ", ")
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *savedataDBRepository) RetrievePlaytime(ctx context.Context, uuid []byte) (int, error) {
	var playtime int
	err := r.db.QueryRowContext(ctx, "SELECT playTime FROM accountStats WHERE uuid = ?", uuid).Scan(&playtime)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return playtime, nil
}

func (r *savedataDBRepository) GetSystemSaveFromS3(ctx context.Context, uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData

	username, err := r.accountRepo.FetchUsernameFromUUID(ctx, uuid)
	if err != nil {
		return system, err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return system, err
	}

	client := s3.NewFromConfig(cfg)

	s3Object := s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_SYSTEM_BUCKET_NAME")),
		Key:    aws.String(username),
	}

	resp, err := client.GetObject(ctx, &s3Object)
	if err != nil {
		return system, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&system)
	if err != nil {
		return system, err
	}

	return system, nil
}
