package storage

import (
	"database/sql"
	"errors"
	"gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	authRedisKeyPref = "auth:"
)

func check(source social.TokenSource, socialID string, tx *sql.Tx) (bool, int64, error) {
	log.Debug("stoarage.check")
	var stmt *sql.Stmt
	var err error
	switch source {
	case social.VKSource:
		stmt, err = tx.Prepare("SELECT `id` FROM `users` u WHERE u.`vk_id` = ?")
	case social.FbSource:
		stmt, err = tx.Prepare("SELECT `id` FROM `users` u WHERE u.`fb_id` = ?")
	default:
		return false, 0, errors.New("invalid token source")
	}
	if err != nil {
		return false, 0, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(socialID)
	if err != nil {
		return false, 0, err
	}
	defer rows.Close()
	registered := rows.Next()
	var userID int64
	if registered {
		err = rows.Scan(&userID)
		if err != nil {
			return true, 0, err
		}
	}
	return registered, userID, nil
}

func register(source social.TokenSource, socialID, name string, tx *sql.Tx) (int64, error) {
	log.Debug("stoarage.register")
	var stmt *sql.Stmt
	var err error
	switch source {
	case social.VKSource:
		stmt, err = tx.Prepare("INSERT INTO `users` (`vk_id`, `name`) VALUES (?, ?)")
	case social.FbSource:
		stmt, err = tx.Prepare("INSERT INTO `users` (`fb_id`, `name`) VALUES (?, ?)")
	default:
		return 0, errors.New("invalid token source")
	}
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(socialID, name)
	if err != nil {
		return 0, err
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

//CheckRegister - function to check if user with given identifier of the given source is registered and if not register. Returns our identifier from the database.
func CheckRegister(source social.TokenSource, socialID, name string, db *sql.DB) (int64, error) {
	log.Debug("storage.CheckRegister")

	var transaction = func(tx *sql.Tx) (int64, error) {
		log.Debug("stoarage.checkregister.transaction")
		registered, userID, err := check(source, socialID, tx)
		if err != nil {
			log.WithError(err).Debug("db check user failed")
			return 0, err
		}
		log.Debug("check user ok")
		if !registered {
			if userID, err = register(source, socialID, name, tx); err != nil {
				log.WithError(err).Debug("db register user failed")
				return 0, err
			}
			log.Debug("register user ok")
		}
		return userID, nil
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	userID, err := transaction(tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	log.WithField("userId", userID).Debug("transaction ok")
	err = tx.Commit()
	if err != nil {
		log.Debug("commit failed")
		return 0, err
	}
	log.Debug("commit ok")
	return userID, nil
}

//GenerateStoreAuthToken - function to generate and save into redis new token (our authorization token) for given user identifier (random string)
func GenerateStoreAuthToken(userID int64, rc *redis.Client) (string, error) {
	log.Debug("stoarage.GenerateStoreAuthToken")
	var authToken string
	for ok := false; !ok; {
		authToken = common.RandStringBytes(20)
		authKey := authRedisKeyPref + authToken
		var err error
		ok, err = rc.SetNX(authKey, userID, time.Hour).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//CheckAuthToken - check given authorization token (from authorization header for example) and return identifier of record in our database
func CheckAuthToken(token string, rc *redis.Client) (int64, error) {
	log.Debug("storage.CheckAuthToken")
	idStr, err := rc.Get(authRedisKeyPref + token).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, &graceful.GracefulUnauthorizedError{}
		}
		return 0, err
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
