package middleware

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pagefaultgames/rogueserver/cache"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/util/logger"
	"github.com/redis/go-redis/v9"
)

// db tier일 때, cache tier일 때 구분하기

// cache tier
func (s *StargateMiddleware) UpdateAllCache(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	var data CombinedSaveData
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	if data.ClientSessionId == "" {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
		return
	}

	var active bool

	active, err = cache.IsActiveSession(uuid, data.ClientSessionId)

	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
		return
	}

	if !active {
		httpError(w, r, fmt.Errorf("session out of date: not active : %s", err), http.StatusBadRequest)
		return
	}

	storedTrainerId, storedSecretId, err := cache.FetchTrainerIds(uuid)
	if err != nil {
		logger.Error("%s", err)
		if errors.Is(err, redis.Nil) {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	if storedTrainerId > 0 || storedSecretId > 0 {
		if data.System.TrainerId != storedTrainerId || data.System.SecretId != storedSecretId {
			httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
			return
		}
	} else {
		err = cache.UpdateTrainerIds(data.System.TrainerId, data.System.SecretId, uuid)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				httpError(w, r, err, http.StatusInternalServerError)
				return
			}
		}
	}

	// cache로 변경
	existingPlaytime, err := cache.RetrievePlaytime(uuid)
	if err != nil && errors.Is(err, redis.Nil) {
		httpError(w, r, fmt.Errorf("failed to retrieve playtime: %s", err), http.StatusInternalServerError)
		return
	} else {
		if data.System.GameStats == nil {
			httpError(w, r, fmt.Errorf("GameStats is nil"), http.StatusBadRequest)
			return
		}
		playtime, ok := data.System.GameStats.(map[string]interface{})["playTime"].(float64)
		if !ok {
			httpError(w, r, fmt.Errorf("no playtime found or invalid type"), http.StatusBadRequest)
			return
		}

		if float64(existingPlaytime) > playtime {
			httpError(w, r, fmt.Errorf("session out of date: existing playtime is greater"), http.StatusBadRequest)
			return
		}
	}

	logger.Info("handleUpdateAll %s %d", uuid, data.SessionSlotId)

	existingSave, err := GetSessionCache(uuid, data.SessionSlotId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		httpError(w, r, fmt.Errorf("failed to retrieve session save data: %s", err), http.StatusInternalServerError)
		return
	} else {
		if existingSave.Seed == data.Session.Seed && existingSave.WaveIndex > data.Session.WaveIndex {
			httpError(w, r, fmt.Errorf("session out of date: existing wave index is greater"), http.StatusBadRequest)
			return
		}
	}

	//logger.Info("Update %s %d %v", uuid, data.SessionSlotId, data.Session)
	err = UpdateCache(uuid, data.SessionSlotId, data.Session)
	if err != nil {
		logger.Error("%v", err)
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = UpdateCache(uuid, 0, data.System)
	if err != nil {
		logger.Error("%v", err)
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// db tier
func (s *StargateMiddleware) UpdateAllDB(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	var data CombinedSaveData
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	if data.ClientSessionId == "" {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
		return
	}

	var active bool

	// db로 변경
	active, err = db.IsActiveSession(uuid, data.ClientSessionId)

	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
		return
	}

	if !active {
		httpError(w, r, fmt.Errorf("session out of date: not active : %s", err), http.StatusBadRequest)
		return
	}

	storedTrainerId, storedSecretId, err := db.FetchTrainerIds(uuid)
	if err != nil {
		logger.Error("%s", err)
		if errors.Is(err, redis.Nil) {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	if storedTrainerId > 0 || storedSecretId > 0 {
		if data.System.TrainerId != storedTrainerId || data.System.SecretId != storedSecretId {
			httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
			return
		}
	} else {
		err = db.UpdateTrainerIds(data.System.TrainerId, data.System.SecretId, uuid)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				httpError(w, r, err, http.StatusInternalServerError)
				return
			}
		}
	}

	existingPlaytime, err := db.RetrievePlaytime(uuid)
	if err != nil && errors.Is(err, redis.Nil) {
		httpError(w, r, fmt.Errorf("failed to retrieve playtime: %s", err), http.StatusInternalServerError)
		return
	} else {
		if data.System.GameStats == nil {
			httpError(w, r, fmt.Errorf("GameStats is nil"), http.StatusBadRequest)
			return
		}
		playtime, ok := data.System.GameStats.(map[string]interface{})["playTime"].(float64)
		if !ok {
			httpError(w, r, fmt.Errorf("no playtime found or invalid type"), http.StatusBadRequest)
			return
		}

		if float64(existingPlaytime) > playtime {
			httpError(w, r, fmt.Errorf("session out of date: existing playtime is greater"), http.StatusBadRequest)
			return
		}
	}

	logger.Info("handleUpdateAll %s %d", uuid, data.SessionSlotId)

	existingSave, err := GetSessionDB(uuid, data.SessionSlotId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		httpError(w, r, fmt.Errorf("failed to retrieve session save data: %s", err), http.StatusInternalServerError)
		return
	} else {
		if existingSave.Seed == data.Session.Seed && existingSave.WaveIndex > data.Session.WaveIndex {
			httpError(w, r, fmt.Errorf("session out of date: existing wave index is greater"), http.StatusBadRequest)
			return
		}
	}

	err = UpdateDB(uuid, data.SessionSlotId, data.Session)
	if err != nil {
		logger.Error("%v", err)
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = UpdateDB(uuid, 0, data.System)
	if err != nil {
		logger.Error("%v", err)
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
