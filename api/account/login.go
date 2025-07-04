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
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/redis/go-redis/v9"
)

type LoginResponse GenericAuthResponse

// /account/login - log into account
func Login(username, password string) (LoginResponse, error) {
	var response LoginResponse

	// 아이디 형식 확인
	if !isValidUsername(username) {
		return response, fmt.Errorf("invalid username")
	}

	// 패스워드 길이 확인
	if len(password) < 6 {
		return response, fmt.Errorf("invalid password")
	}

	// 비밀번호 인증을 위해 필요한 데이터 해시키, 솔트를 데이터베이스에서 가져오기
	key, salt, err := db.FetchAccountKeySaltFromUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, fmt.Errorf("account doesn't exist")
		}

		return response, err
	}

	// 해시 값이랑 내 패스워드, 솔트 값으로 확인
	if !bytes.Equal(key, deriveArgon2IDKey([]byte(password), salt)) {
		return response, fmt.Errorf("password doesn't match")
	}

	// 일치하는 경우 토큰 생성
	response.Token, err = GenerateTokenForUsername(username)

	if err != nil {
		return response, fmt.Errorf("failed to generate token: %s", err)
	}

	// 토큰 반환
	return response, nil
}

func GenerateTokenForUsername(username string) (string, error) {
	token := make([]byte, TokenSize)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %s", err)
	}

	// db에서 uuid 가져오기
	uuid, err := db.FetchUUIDFromUsername(username)
	if err != nil {
		return "", fmt.Errorf("failed to get uuid")
	}

	// uuid와토큰으로 Cache 추가
	// token / uuid
	err = cache.StoreSessionToken(uuid, token)
	if err != nil {
		return "", fmt.Errorf("failed to store token")
	}

	// 유저가 로그인한 것이기 때문에 Cache에 Userdata가 있는지 확인
	err = cache.IsValidCacheData(uuid)

	// 데이터가 없는 경우 넘어가기
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			// 데이터가 있으면 return
			return "", fmt.Errorf("")
		}
	}

	// cache에 해당 uuid를 가진 데이터가 없는 경우
	// db에서 로그인 유저 정보 가져와서 cache에 저장
	// uuid / userData
	accountData, err := db.GetAccountFromDB(uuid)

	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	// account 정보 cache로 가져오기
	cache.CacheAccountInRedis(accountData)

	// db에서 로그인 유저 통계 정보 가져와서 cache에 저장
	accountStatsData, err := db.GetAccountStatsFromDB(uuid)
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	// accountStats 정보 cache로 가져오기
	cache.CacheAccountStatsInRedis(accountStatsData)

	// 유저 아이디와 토큰값으로 세션 정보 저장 -> DB에 굳이 할 필요가 없어짐
	// err = db.AddAccountSession(username, token)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to add account session")
	// }

	return base64.StdEncoding.EncodeToString(token), nil
}
