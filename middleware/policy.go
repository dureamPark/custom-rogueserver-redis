package middleware

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/redis/go-redis/v9"
)

func calculateFinalScore(realtimeScore float64, profileScore float64, sessionDuration time.Duration) float64 {
	T_full_confidence := 10 * time.Minute // 10분이면 신뢰도 100%

	// 1. 신뢰도(C) 계산
	confidence := float64(sessionDuration) / float64(T_full_confidence)
	if confidence > 1.0 {
		confidence = 1.0
	}

	// 2. 최종 점수 계산
	finalScore := (confidence * realtimeScore) + ((1.0 - confidence) * profileScore)

	return finalScore
}

func UpdateFinalScore(context Context.context, uuid []byte, score float64) {
	// active score 증가
	activityKey := "user_activity_score:" + base64.StdEncoding.EncodeToString(uuid)
	if err := cache.Rdb.Incr(context.Background(), activityKey).Err(); err != nil {
		log.Printf("[Stargate] ERROR: Failed to update activity for %s: %v", uuid, err)
	}
	// ------------------------------------------
}

func updateRealtimeScore_WithDecay(context Context.context, uuid []byte, profileScore float64, realtimeScore float64, sessionDuration time.Duration) {

	encodedUUID := base64.StdEncoding.EncodeToString(uuid)

	// 1. Redis에서 현재 점수와 마지막 업데이트 시간을 가져온다.
	currentScore := cache.Rdb.Get(context, "user_realtime_score:"+encodedUUID)
	lastTimestamp := cache.Rdb.Get(context, "user_last_timestamp:"+encodedUUID)

	currentTime := time.Now().Unix()

	// 2. 시간 경과에 따른 점수 '감쇠' 계산
	var decayedScore float64
	if lastTimestamp > 0 {
		timeDelta := float64(currentTime - lastTimestamp)
		lambda := 0.005 // 감쇠 상수

		// 지수적 감쇠 공식: Score_new = Score_old * e^(-λ * Δt)
		decayFactor := math.Exp(-lambda * timeDelta)
		decayedScore = currentScore * decayFactor
	} else {
		// 첫 활동인 경우
		decayedScore = 0.0
	}

	// 3. 이번 활동(가치=1) 점수를 더한다.
	newScore := decayedScore + 1.0

	finalScore := calculateFinalScore(newScore, profileScore, 0)

	// 4. 갱신된 점수와 현재 시간을 다시 Redis에 저장한다.
	cache.Rdb.Set(context, "user_realtime_score:"+encodedUUID, newScore)
	cache.Rdb.Set(context, "user_last_timestamp:"+encodedUUID, currentTime)

	Stargate.redisClient.Set(r.Context(), key, fmt.Sprintf("%f", finalScore), redis.KeepTTL)

}
