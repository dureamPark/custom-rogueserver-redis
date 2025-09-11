package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/pagefaultgames/rogueserver/api/account"
	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/defs"
	"github.com/pagefaultgames/rogueserver/util/logger"
	"github.com/redis/go-redis/v9"
)

// ... ContextKey 정의 ...

type StargateMiddleware struct {
	redisClient  *redis.Client // ★ Redis 클라이언트를 직접 의존성으로 가짐
	cacheHandler http.Handler
	dbHandler    http.Handler
}

type CombinedSaveData struct {
	System          defs.SystemSaveData  `json:"system"`
	Session         defs.SessionSaveData `json:"session"`
	SessionSlotId   int                  `json:"sessionSlotId"`
	ClientSessionId string               `json:"clientSessionId"`
}

func NewStargateMiddleware(client *redis.Client, cache http.Handler, db http.Handler) *StargateMiddleware {
	return &StargateMiddleware{
		redisClient:  client,
		cacheHandler: cache,
		dbHandler:    db,
	}
}

func (s *StargateMiddleware) UpdateAll(w http.ResponseWriter, r *http.Request) {

	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	// --- 데이터 접근 로직이 미들웨어에 직접 존재 ---
	key := "user_tier:" + base64.StdEncoding.EncodeToString(uuid)
	tier, err := s.redisClient.Get(r.Context(), key).Result()
	if err != nil && err != redis.Nil {
		log.Printf("[Stargate] CRITICAL: Failed to get user tier for %s: %v", uuid, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	go func(uid string) {
		activityKey := "user_activity_score:" + base64.StdEncoding.EncodeToString(uuid)
		if err := s.redisClient.Incr(context.Background(), activityKey).Err(); err != nil {
			log.Printf("[Stargate] ERROR: Failed to update activity for %s: %v", uid, err)
		}
	}(string(uuid))
	// ------------------------------------------

	switch tier {
	case "CACHE":
		log.Printf("[Stargate] User %s routed to CACHE.", uuid)
		s.UpdateAllCache(w, r)
	default:
		log.Printf("[Stargate] User %s routed to DB (Tier: '%s').", uuid, tier)
		s.UpdateAllDB(w, r)
	}
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
	uuid, err := cache.FetchSessionToken(token)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			logger.Error("token : %s\n", base64.StdEncoding.EncodeToString(token))
			return nil, nil, fmt.Errorf("redis GET error: %w", err)
		}
	}

	return token, uuid, nil
}

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
