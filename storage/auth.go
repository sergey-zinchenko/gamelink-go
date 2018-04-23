package storage

import (
	"gamelink-go/graceful"
	"github.com/go-redis/redis"
	"time"
	"database/sql"
	"gamelink-go/social"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
)

const (
	letters          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	authRedisKeyPref = "auth:"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func vkCheck(vkId string, tx *sql.Tx) (bool, int64, *graceful.Error) {
	log.Debug("stoarage.vkCheck")
	stmt, err := tx.Prepare("SELECT `id` FROM `users` u WHERE u.`vk_id` = ?")
	if err != nil {
		return false, 0, graceful.NewMySqlError(err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.Query(vkId)
	if err != nil {
		return false, 0, graceful.NewMySqlError(err.Error())
	}
	defer rows.Close()
	registered := rows.Next()
	var userId int64
	if registered {
		err = rows.Scan(&userId)
		if err != nil {
			return true, 0, graceful.NewMySqlError(err.Error())
		}
	}
	return registered, userId, nil
}

func vkRegister(vkId string, name string, tx *sql.Tx) (int64, *graceful.Error) {
	log.Debug("stoarage.vkRegister")
	stmt, err := tx.Prepare("INSERT INTO `users` (`vk_id`, `name`) VALUES (?, ?)")
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	defer stmt.Close()
	res, err := stmt.Exec(vkId, name)
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	userId, err := res.LastInsertId()
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	return userId, nil
}

func fbCheck(vkId string, tx *sql.Tx) (bool, int64, *graceful.Error) {
	log.Debug("stoarage.fbCheck")
	stmt, err := tx.Prepare("SELECT `id` FROM `users` u WHERE u.`fb_id` = ?")
	if err != nil {
		return false, 0, graceful.NewMySqlError(err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.Query(vkId)
	if err != nil {
		return false, 0, graceful.NewMySqlError(err.Error())
	}
	defer rows.Close()
	registered := rows.Next()
	var userId int64
	if registered {
		err = rows.Scan(&userId)
		if err != nil {
			return true, 0, graceful.NewMySqlError(err.Error())
		}
	}
	return registered, userId, nil
}

func fbRegister(vkId string, name string, tx *sql.Tx) (int64, *graceful.Error) {
	log.Debug("stoarage.fbRegister")
	stmt, err := tx.Prepare("INSERT INTO `users` (`fb_id`, `name`) VALUES (?, ?)")
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	defer stmt.Close()
	res, err := stmt.Exec(vkId, name)
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	userId, err := res.LastInsertId()
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	return userId, nil
}

func VkCheckRegister(token string, db *sql.DB) (int64, *graceful.Error) {
	log.Debug("stoarage.VkCheckRegister")
	vkId, name, err := social.NewVkToken(token).GetUserInfo()
	if err != nil {
		log.WithError(err).Debug("vk user info failed")
		return 0, err
	}
	log.Debug("vk user info ok")
	var transaction = func(tx *sql.Tx) (int64, *graceful.Error) {
		log.Debug("auth.checkregister.transaction")
		registered, userId, err := vkCheck(vkId, tx)
		if err != nil {
			log.WithError(err).Debug("db check user failed")
			return 0, err
		}
		log.Debug("db check user ok")
		if !registered {
			if userId, err = vkRegister(vkId, name, tx); err != nil {
				log.WithError(err).Debug("db register user failed")
				return 0, err
			}
			log.Debug("db register user ok")
		}
		return userId, nil
	}
	tx, e := db.Begin()
	if e != nil {
		return 0, graceful.NewMySqlError(e.Error())
	}
	userId, err := transaction(tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	log.WithField("userId", userId).Debug("transaction ok")
	e = tx.Commit()
	if e != nil {
		log.Debug("commit failed")
		return 0, graceful.NewMySqlError(e.Error())
	}
	log.Debug("commit ok")
	return userId, nil
}

func FbCheckRegister(token string, db *sql.DB) (int64, *graceful.Error) {
	log.Debug("stoarage.FbCheckRegister")
	fbId, name, err := social.NewFbToken(token).GetUserInfo()
	if err != nil {
		log.WithError(err).Debug("fb user info failed")
		return 0, err
	}
	log.Debug("fb user info ok")
	var transaction = func(tx *sql.Tx) (int64, *graceful.Error) {
		log.Debug("auth.checkregister.transaction")
		registered, userId, err := fbCheck(fbId, tx)
		if err != nil {
			log.WithError(err).Debug("db check user failed")
			return 0, err
		}
		log.Debug("db check user ok")
		if !registered {
			if userId, err = fbRegister(fbId, name, tx); err != nil {
				log.WithError(err).Debug("db register user failed")
				return 0, err
			}
			log.Debug("db register user ok")
		}
		return userId, nil
	}
	tx, e := db.Begin()
	if e != nil {
		return 0, graceful.NewMySqlError(e.Error())
	}
	userId, err := transaction(tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	log.WithField("userId", userId).Debug("transaction ok")
	e = tx.Commit()
	if e != nil {
		log.Debug("commit failed")
		return 0, graceful.NewMySqlError(e.Error())
	}
	log.Debug("commit ok")
	return userId, nil
}

func GenerateStoreAuthToken(userId int64, rc *redis.Client) (string, *graceful.Error) {
	log.Debug("stoarage.GenerateStoreAuthToken")
	authToken := RandStringBytes(20)
	authKey := authRedisKeyPref + authToken
	for ok := false; !ok; {
		var err error
		ok, err = rc.SetNX(authKey, userId, time.Hour).Result()
		if err != nil {
			return "", graceful.NewRedisError(err.Error())
		}
	}
	return authToken, nil
}

func CheckAuthToken(token string, rc *redis.Client) (uint64, *graceful.Error) {
	log.Debug("storage.CheckAuthToken")
	idStr, err := rc.Get(authRedisKeyPref + token).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, graceful.NewInvalidError("key does not exists")
		} else {
			return 0, graceful.NewRedisError(err.Error())
		}
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, graceful.NewParsingError(err.Error())
	}
	return id, nil
}