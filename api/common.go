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

package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pagefaultgames/rogueserver/api/account"
	"github.com/pagefaultgames/rogueserver/api/daily"
	"github.com/pagefaultgames/rogueserver/repository"
	"github.com/pagefaultgames/rogueserver/util/logger"
	"github.com/redis/go-redis/v9"
	//"github.com/pagefaultgames/rogueserver/db"
)

const (
	sessionUUIDKeyFmt = "session:uuid:%s"
	sessionTTL        = 7 * 24 * time.Hour
)

// InitWithRepositories: 레포지토리 주입 및 API 핸들러 등록
func InitWithRepositories(mux *http.ServeMux) error {

	err := scheduleStatRefresh()
	if err != nil {
		return err
	}

	err = daily.Init()
	if err != nil {
		return err
	}

	// account
	mux.HandleFunc("GET /account/info", handleAccountInfo)
	mux.HandleFunc("POST /account/register", handleAccountRegister)
	mux.HandleFunc("POST /account/login", handleAccountLogin)
	mux.HandleFunc("POST /account/changepw", handleAccountChangePW)
	mux.HandleFunc("GET /account/logout", handleAccountLogout)

	// game
	mux.HandleFunc("GET /game/titlestats", handleGameTitleStats)
	mux.HandleFunc("GET /game/classicsessioncount", handleGameClassicSessionCount)

	// savedata
	mux.HandleFunc("/savedata/session/{action}", handleSession)
	mux.HandleFunc("/savedata/system/{action}", handleSystem)

	// new session
	mux.HandleFunc("POST /savedata/updateall", handleUpdateAll)

	// daily
	mux.HandleFunc("GET /daily/seed", handleDailySeed)
	mux.HandleFunc("GET /daily/rankings", handleDailyRankings)
	mux.HandleFunc("GET /daily/rankingpagecount", handleDailyRankingPageCount)

	// auth
	mux.HandleFunc("/auth/{provider}/callback", handleProviderCallback)
	mux.HandleFunc("/auth/{provider}/logout", handleProviderLogout)

	// admin
	mux.HandleFunc("POST /admin/account/discordLink", handleAdminDiscordLink)
	mux.HandleFunc("POST /admin/account/discordUnlink", handleAdminDiscordUnlink)
	mux.HandleFunc("POST /admin/account/googleLink", handleAdminGoogleLink)
	mux.HandleFunc("POST /admin/account/googleUnlink", handleAdminGoogleUnlink)
	mux.HandleFunc("GET /admin/account/adminSearch", handleAdminSearch)

	return nil
}

func tokenFromRequest(r *http.Request) ([]byte, error) {
	if r.Header.Get("Authorization") == "" {
		return nil, fmt.Errorf("missing token")
	}

	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %s", err)
	}

	if len(token) != account.TokenSize {
		return nil, fmt.Errorf("invalid token length: got %d, expected %d", len(token), account.TokenSize)
	}

	return token, nil
}

func uuidFromRequest(r *http.Request) ([]byte, error) {
	_, uuid, err := tokenAndUuidFromRequest(r)
	if err != nil {
		return nil, err
	}

	return uuid, nil
}

func tokenAndUuidFromRequest(r *http.Request) ([]byte, []byte, error) {
	// 토큰은 cache에서만 확인하기

	// 1) Header에서 토큰 추출
	token, err := tokenFromRequest(r)
	if err != nil {
		return nil, nil, err
	}

	// 2) Redis 캐시 조회 (token→uuid)
	logger.Info("token : %s\n", base64.StdEncoding.EncodeToString(token))
	uuid, err := repository.Repos.Account.FetchUUIDFromToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Error("token : %s\n", base64.StdEncoding.EncodeToString(token))
			return nil, nil, fmt.Errorf("redis GET error: %w", err)
		}
	}

	// // 3) cache miss → DB 조회
	// uuid, err := db.FetchUUIDFromToken(token)
	// if err != nil {
	//     return nil, nil, fmt.Errorf("failed to validate token: %w", err)
	// }
	// // 4) 캐시에 SET
	// cache.Rdb.Set(cache.Ctx, cacheKey, string(uuid), sessionTTL)

	return token, uuid, nil
}

/*
func tokenAndUuidFromRequest(r *http.Request) ([]byte, []byte, error) {
    token, err := tokenFromRequest(r)
    if err != nil {
        return nil, nil, err
    }
    // 1) 캐시로 UUID 조회
    tokStr := base64.StdEncoding.EncodeToString(token)
    if u, err := cache.Rdb.Get(cache.Ctx, fmt.Sprintf(sessionUUIDKeyFmt, tokStr)).Result(); err == nil {
        return token, []byte(u), nil
    } else if err != redis.Nil {
        return nil, nil, fmt.Errorf("redis GET uuid error: %w", err)
    }
    // 2) 캐시 miss → DB 조회
    uuid, err := db.FetchUUIDFromToken(token)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to validate token: %s", err)
    }
    // 3) 캐시에 token→uuid 매핑 저장 (optional, TTL=1주일)
    cache.Rdb.Set(cache.Ctx, fmt.Sprintf(sessionUUIDKeyFmt, tokStr), string(uuid), sessionTTL)
    return token, uuid, nil
}
*/

/*func tokenAndUuidFromRequest(r *http.Request) ([]byte, []byte, error) {
	token, err := tokenFromRequest(r)
	if err != nil {
		return nil, nil, err
	}

	uuid, err := db.FetchUUIDFromToken(token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate token: %s", err)
	}

	return token, uuid, nil
}*/

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	logger.Error("%s: %s\n", r.URL.Path, err)
	http.Error(w, err.Error(), code)
}

func writeJSON(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}
