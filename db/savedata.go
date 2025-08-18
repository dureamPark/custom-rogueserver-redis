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

package db

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"os"
	"strings"

	"github.com/pagefaultgames/rogueserver/util/logger"

	"github.com/klauspost/compress/zstd"
	"github.com/pagefaultgames/rogueserver/defs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TryAddSeedCompletion(uuid []byte, seed string, mode int) (bool, error) {
	var count int
	err := handle.QueryRow("SELECT COUNT(*) FROM dailyRunCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	} else if count > 0 {
		return false, nil
	}

	_, err = handle.Exec("INSERT INTO dailyRunCompletions (uuid, seed, mode, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP())", uuid, seed, mode)
	if err != nil {
		return false, err
	}

	return true, nil
}

func ReadSeedCompleted(uuid []byte, seed string) (bool, error) {
	var count int
	err := handle.QueryRow("SELECT COUNT(*) FROM dailyRunCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func ReadSystemSaveData(uuid []byte) (defs.SystemSaveData, error) {
	//logger.Info("ReadSystemSaveData %s", uuid);
	var system defs.SystemSaveData

	var data []byte
	err := handle.QueryRow("SELECT data FROM systemSaveData WHERE uuid = ?", uuid).Scan(&data)
	if err != nil {
		logger.Info("Not Find Data")
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

func StoreSystemSaveData(uuid []byte, data defs.SystemSaveData) error {
	//logger.Info("StoreSystemSaveData %s %v", uuid, data)

	buf := new(bytes.Buffer)

	zw, err := zstd.NewWriter(buf)
	if err != nil {
		return err
	}

	//defer zw.Close()

	err = gob.NewEncoder(zw).Encode(data)
	if err != nil {
		logger.Error("Encoding Error: %v", err)
		return err
	}

	zw.Flush()
	logger.Info("✅ Data flushed to buffer!")

	err = zw.Close()
	if err != nil {
		logger.Error("❌ Failed to close ZSTD writer: %v", err)
	} else {
		logger.Info("✅ ZSTD writer closed successfully")
	}

	logger.Info("Compressed Data Length: %d", len(buf.Bytes()))
	//logger.Info("Compressed Data Content: %v", buf.Bytes())

	_, err = handle.Exec("REPLACE INTO systemSaveData (uuid, data, timestamp) VALUES (?, ?, UTC_TIMESTAMP())", uuid, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func StoreSystemSaveDataS3(uuid []byte, data defs.SystemSaveData) error {
	cfg, _ := config.LoadDefaultConfig(context.TODO())

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL_S3"))
	})

	username, err := FetchUsernameFromUUID(uuid)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	err = json.NewEncoder(buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("S3_SYSTEM_BUCKET_NAME")),
		Key:    aws.String(username),
		Body:   buf,
	})
	if err != nil {
		return err
	}

	return nil
}

func DeleteSystemSaveData(uuid []byte) error {
	_, err := handle.Exec("DELETE FROM systemSaveData WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func ReadSessionSaveData(uuid []byte, slot int) (defs.SessionSaveData, error) {
	logger.Info("ReadSessionSaveData %s %d", uuid, slot)
	var session defs.SessionSaveData

	var data []byte
	err := handle.QueryRow("SELECT data FROM sessionSaveData WHERE uuid = ? AND slot = ?", uuid, slot).Scan(&data)
	if err != nil {
		return session, err
	}

	zr, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return session, err
	}

	//defer zr.Close()

	err = gob.NewDecoder(zr).Decode(&session)
	if err != nil {
		return session, err
	}

	zr.Close()

	return session, nil
}

func GetLatestSessionSaveDataSlot(uuid []byte) (int, error) {
	var slot int
	logger.Info("getlatestsessionsavedataslot")
	err := handle.QueryRow("SELECT slot FROM sessionSaveData WHERE uuid = ? ORDER BY timestamp DESC, slot ASC LIMIT 1", uuid).Scan(&slot)
	if err != nil {
		return -1, err
	}

	return slot, nil
}

func StoreSessionSaveData(uuid []byte, data defs.SessionSaveData, slot int) error {
	logger.Info("StoreSessionSaveData %s %v %d", uuid, data, slot)
	buf := new(bytes.Buffer)

	zw, err := zstd.NewWriter(buf)
	if err != nil {
		return err
	}

	//defer zw.Close()

	err = gob.NewEncoder(zw).Encode(data)
	if err != nil {
		logger.Error("Encoding Error: %v", err)
		return err
	}

	zw.Flush()
	zw.Close()

	_, err = handle.Exec("REPLACE INTO sessionSaveData (uuid, slot, data, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP())", uuid, slot, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func DeleteSessionSaveData(uuid []byte, slot int) error {
	_, err := handle.Exec("DELETE FROM sessionSaveData WHERE uuid = ? AND slot = ?", uuid, slot)
	if err != nil {
		return err
	}

	return nil
}

// StoreSessionSaveDataBulk는 주어진 UUID에 대한 여러 세션 데이터를 DB에 일괄 업데이트합니다.
// 함수 시그니처를 호출하는 쪽의 맥락에 맞게 재정의했습니다.
// (예: bulk 처리를 위해 uuid 슬라이스와 sessionsDataMap 슬라이스를 받도록)
func StoreSessionSaveDataBulk(ctx context.Context, uuids [][]byte, sessionsDataMapList []defs.SessionSaveData, slot int) error {
	logger.Info("StoreSessionSaveDataBulk processing a batch...")

	// 단일 트랜잭션 시작: 모든 업데이트가 성공하거나 모두 실패하도록 보장
	tx, err := handle.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// defer로 롤백을 호출하여, 에러 발생 시 자동 롤백되도록 함
	defer tx.Rollback()

	// 쿼리 파라미터들을 담을 슬라이스
	var args []interface{}
	// BULK INSERT의 VALUES 부분을 생성하기 위한 슬라이스
	var valuePlaceholders []string

	// 데이터들을 순회하며 파라미터와 플레이스홀더를 준비합니다.
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

		// UUID, sessionID, data, timestamp에 대한 플레이스홀더를 추가합니다.
		// (?, ?, ?, UTC_TIMESTAMP())는 한 행에 대한 플레이스홀더입니다.
		valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, UTC_TIMESTAMP())")

		// 쿼리 실행에 필요한 파라미터들을 순서대로 슬라이스에 추가합니다.
		args = append(args, uuid, slot, buf.Bytes())
	}

	if len(args) == 0 {
		return nil // 처리할 데이터가 없으면 종료
	}

	// 동적으로 VALUES 부분을 조합하여 하나의 쿼리를 생성합니다.
	// REPLACE INTO는 기존 행을 삭제 후 재삽입하므로, INSERT ... ON DUPLICATE KEY UPDATE가 더 효율적입니다.
	query := "REPLACE INTO sessionSaveData (uuid, slot, data, timestamp) VALUES " + strings.Join(valuePlaceholders, ", ")

	// 하나의 트랜잭션으로 모든 데이터를 한 번에 실행합니다.
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// 모든 작업이 성공했으므로 트랜잭션을 커밋합니다.
	return tx.Commit()
}

func StoreSystemSaveDataBulk(ctx context.Context, uuids [][]byte, systemsDataList []defs.SystemSaveData) error {
	logger.Info("StoreSystemSaveDataBulk processing a batch...")

	// 단일 트랜잭션 시작: 모든 업데이트가 성공하거나 모두 실패하도록 보장
	tx, err := handle.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// defer로 롤백을 호출하여, 에러 발생 시 자동 롤백되도록 함
	defer tx.Rollback()

	// 쿼리 파라미터들을 담을 슬라이스
	var args []interface{}
	// BULK INSERT의 VALUES 부분을 생성하기 위한 슬라이스
	var valuePlaceholders []string

	// 데이터들을 순회하며 파라미터와 플레이스홀더를 준비합니다.
	for i, data := range systemsDataList {
		uuid := uuids[i]

		// gob와 zstd를 사용하여 데이터를 직렬화 및 압축합니다.
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

		// UUID, data, timestamp에 대한 플레이스홀더를 추가합니다.
		// (?, ?, UTC_TIMESTAMP())는 한 행에 대한 플레이스홀더입니다.
		valuePlaceholders = append(valuePlaceholders, "(?, ?, UTC_TIMESTAMP())")

		// 쿼리 실행에 필요한 파라미터들을 순서대로 슬라이스에 추가합니다.
		args = append(args, uuid, buf.Bytes())
	}

	if len(args) == 0 {
		return nil // 처리할 데이터가 없으면 종료
	}

	// 동적으로 VALUES 부분을 조합하여 하나의 쿼리를 생성합니다.
	query := "REPLACE INTO systemSaveData (uuid, data, timestamp) VALUES " + strings.Join(valuePlaceholders, ", ")

	// 하나의 트랜잭션으로 모든 데이터를 한 번에 실행합니다.
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// 모든 작업이 성공했으므로 트랜잭션을 커밋합니다.
	return tx.Commit()
}

func RetrievePlaytime(uuid []byte) (int, error) {
	var playtime int
	err := handle.QueryRow("SELECT playTime FROM accountStats WHERE uuid = ?", uuid).Scan(&playtime)
	if err != nil {
		return 0, err
	}

	return playtime, nil
}

func GetSystemSaveFromS3(uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData

	username, err := FetchUsernameFromUUID(uuid)
	if err != nil {
		return system, err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return system, err
	}

	client := s3.NewFromConfig(cfg)

	s3Object := s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_SYSTEM_BUCKET_NAME")),
		Key:    aws.String(username),
	}

	resp, err := client.GetObject(context.TODO(), &s3Object)
	if err != nil {
		return system, err
	}

	err = json.NewDecoder(resp.Body).Decode(&system)
	if err != nil {
		return system, err
	}

	return system, nil
}
