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
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pagefaultgames/rogueserver/repository"
	"github.com/pagefaultgames/rogueserver/util/logger"
)

type LoginResponse GenericAuthResponse

// /account/login - log into account
func Login(ctx context.Context, username, password string) (LoginResponse, error) {
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
	key, err := repository.Repos.Account.FetchAccountKeySaltFromUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, fmt.Errorf("account doesn't exist")
		}

		return response, err
	}

	// 해시 값이랑 내 패스워드, 솔트 값으로 확인
	passwordbyte := make([]byte, 32)
	copy(passwordbyte, password)
	if !bytes.Equal(key, passwordbyte) {
		logger.Error("password doesn't match key: %v, password: %v", key, passwordbyte)
		return response, fmt.Errorf("password doesn't match")
	}

	// 일치하는 경우 토큰 생성
	response.Token, err = GenerateTokenForUsername(ctx, username)

	if err != nil {
		return response, fmt.Errorf("failed to generate token: %s", err)
	}

	// 토큰 반환
	return response, nil
}

func GenerateTokenForUsername(ctx context.Context, username string) (string, error) {
	token := make([]byte, TokenSize)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %s", err)
	}

	err = repository.Repos.Account.AddAccountSession(ctx, username, token)
	if err != nil {
		return "", fmt.Errorf("failed to add account session")
	}

	return base64.StdEncoding.EncodeToString(token), nil
}
