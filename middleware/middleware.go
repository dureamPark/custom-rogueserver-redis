package middleware

import (
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
	redisClient *redis.Client // ★ Redis 클라이언트를 직접 의존성으로 가짐
}

var (
	Stargate *StargateMiddleware
)

type CombinedSaveData struct {
	System          defs.SystemSaveData  `json:"system"`
	Session         defs.SessionSaveData `json:"session"`
	SessionSlotId   int                  `json:"sessionSlotId"`
	ClientSessionId string               `json:"clientSessionId"`
}

func NewStargateMiddleware(client *redis.Client) {
	Stargate.redisClient = client
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
		tier = "CACHE"
		cache.SetUserTier(r.Context(), uuid, "CACHE")

		// cache에 없으니까 db에서 가져와서 점수 계산 후 넣기
		profileScore, err := GetProfileScore()

		if err != nil {
			return
		}

		updateRealtimeScore_WithDecay(uuid, 0, profileScore, 0)

		// // cache에 없으니까 db에서 가져와서 점수 계산 후 넣기
		// profileScore, err := GetProfileScore()
		// finalScore := calculateFinalScore(0, profileScore, 0)
		// Stargate.redisClient.Set(r.Context(), key, fmt.Sprintf("%f", finalScore), redis.KeepTTL)

		// --- tier가 설정되지 않은 경우 (세션 시작 등), 초기 배정 로직 실행 ---
		//log.Printf("[Stargate] Tier not set for user %s. Performing initial assignment...", uuid)

		// // 4. 초기 배정을 위해 프로필과 시스템 임계치 조회
		// profileScore, err := cache.GetAccountProfileScore(r.Context(), uuid)
		// if err != nil {
		// 	log.Printf("[Stargate] ERROR: Failed to get profile for %s: %v", uuid, err)
		// 	return
		// }

		// // Oracle이 미리 계산해 둔 시스템 커트라인 조회
		// systemThreshold, _ := cache.GetSystemThreshold(r.Context()) // PolicyStorer에 추가 필요

		// // 5. 초기 우선순위 점수 '계산' (메모리 상에서만)
		// initialScore := calculateFinalScore(0, profileScore, 0)

		// // 6. '결정' 및 '결과 저장'
		// if initialScore >= systemThreshold {
		// 	log.Printf("[Stargate] Assigning user %s to CACHE tier.", uuid)

		// 	// 결정된 '상태'("CACHE")를 Redis에 저장
		// 	cache.SetUserTier(r.Context(), uuid, "CACHE")
		// } else {
		// 	log.Printf("[Stargate] Assigning user %s to DB tier.", uuid)

		// 	// 결정된 '상태'("DB")를 Redis에 저장
		// 	cache.SetUserTier(r.Context(), uuid, "DB")
		// }

		// log.Printf("[Stargate] CRITICAL: Failed to get user tier for %s: %v", uuid, err)
		// http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

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
