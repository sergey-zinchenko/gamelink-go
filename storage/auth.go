package storage

import (
	"gamelink-go/graceful"
	"github.com/go-redis/redis"
	"time"
	"database/sql"
	"gamelink-go/social"
	log "github.com/sirupsen/logrus"
	"strconv"
	"gamelink-go/common"
)

const (
	authRedisKeyPref = "auth:"
)

func check(source social.TokenSource, socialId string, tx *sql.Tx) (bool, int64, *graceful.Error) {
	log.Debug("stoarage.check")
	var stmt *sql.Stmt
	var err error
	switch source {
	case social.VKSource:
		stmt, err = tx.Prepare("SELECT `id` FROM `users` u WHERE u.`vk_id` = ?")
	case social.FbSource:
		stmt, err = tx.Prepare("SELECT `id` FROM `users` u WHERE u.`fb_id` = ?")
	default:
		return false, 0, graceful.NewInvalidError("invalid token source")
	}
	if err != nil {
		return false, 0, graceful.NewMySqlError(err.Error())
	}
	defer stmt.Close()
	rows, err := stmt.Query(socialId)
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

func register(source social.TokenSource, socialId, name string, tx *sql.Tx) (int64, *graceful.Error) {
	log.Debug("stoarage.register")
	var stmt *sql.Stmt
	var err error
	switch source {
	case social.VKSource:
		stmt, err = tx.Prepare("INSERT INTO `users` (`vk_id`, `name`) VALUES (?, ?)")
	case social.FbSource:
		stmt, err = tx.Prepare("INSERT INTO `users` (`fb_id`, `name`) VALUES (?, ?)")
	default:
		return 0, graceful.NewInvalidError("invalid token source")
	}
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	defer stmt.Close()
	res, err := stmt.Exec(socialId, name)
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	userId, err := res.LastInsertId()
	if err != nil {
		return 0, graceful.NewMySqlError(err.Error())
	}
	return userId, nil
}

func CheckRegister(source social.TokenSource, socialId, name string, db *sql.DB) (int64, *graceful.Error) {
	log.Debug("stoarage.CheckRegister")

	var transaction = func(tx *sql.Tx) (int64, *graceful.Error) {
		log.Debug("stoarage.checkregister.transaction")
		registered, userId, err := check(source, socialId, tx)
		if err != nil {
			log.WithError(err).Debug("db check user failed")
			return 0, err
		}
		log.Debug("check user ok")
		if !registered {
			if userId, err = register(source, socialId, name, tx); err != nil {
				log.WithError(err).Debug("db register user failed")
				return 0, err
			}
			log.Debug("register user ok")
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
	authToken := common.RandStringBytes(20)
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

func CheckAuthToken(token string, rc *redis.Client) (int64, *graceful.Error) {
	log.Debug("storage.CheckAuthToken")
	idStr, err := rc.Get(authRedisKeyPref + token).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, graceful.NewNotFoundError("key does not exists")
		} else {
			return 0, graceful.NewRedisError(err.Error())
		}
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, graceful.NewParsingError(err.Error())
	}
	return id, nil
}
