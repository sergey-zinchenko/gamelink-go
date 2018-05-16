package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

const (
	authRedisKeyPref = "auth:"
)

func check(socialID social.ThirdPartyID, tx *sql.Tx) (bool, int64, error) {
	var userID int64
	queryString := fmt.Sprintf("SELECT `id` FROM `users` u WHERE u.`%s` = ?", socialID.Name())
	err := tx.QueryRow(queryString, socialID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}
	return true, userID, nil
}

func register(socialID social.ThirdPartyID, name string, tx *sql.Tx) (int64, error) {
	b, err := json.Marshal(C.J{socialID.Name(): socialID, "name": name})
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec("INSERT INTO `users` (`data`) VALUES (?)", b)
	if err != nil {
		return 0, err
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

//AuthorizedUser - function to check our own authorization token from header. Returns user or nil if not valid token.
func (dbs DBS) AuthorizedUser(token string) (*User, error) {
	if dbs.rc == nil {
		return nil, errors.New("databases not initialized")
	}
	idStr, err := dbs.rc.Get(authRedisKeyPref + token).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, &graceful.UnauthorizedError{Message: "key does not exists in redis"}
		}
		return nil, err
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return &User{id, &dbs}, nil
}

//ThirdPartyUser - function to login or register user using his third party token
func (dbs DBS) ThirdPartyUser(token social.ThirdPartyToken) (*User, error) {
	if dbs.mySQL == nil {
		return nil, errors.New("databases not initialized")
	}
	var transaction = func(socialID social.ThirdPartyID, name string, friendIds []social.ThirdPartyID, tx *sql.Tx) (int64, error) {
		registered, userID, err := check(socialID, tx)
		if err != nil {
			return 0, err
		}
		if !registered {
			if userID, err = register(socialID, name, tx); err != nil {
				return 0, err
			}
		}
		err = dbs.SyncFriends(friendIds, userID, tx)
		if err != nil {
			return 0, err
		}
		return userID, nil
	}
	socialID, name, friendsIds, err := token.UserInfo()
	if err != nil {
		return nil, err
	}
	tx, err := dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	userID, err := transaction(socialID, name, friendsIds, tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &User{userID, &dbs}, nil
}

//AuthToken - Function to generate and store auth token in rc.
func (u User) AuthToken() (string, error) {
	if u.dbs.rc == nil {
		return "", errors.New("databases not initialized")
	}
	var authToken string
	for ok := false; !ok; {
		authToken = C.RandStringBytes(20)
		authKey := authRedisKeyPref + authToken
		var err error
		ok, err = u.dbs.rc.SetNX(authKey, u.ID(), time.Hour).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//SyncFriends - add user friends to table friends
func (dbs DBS) SyncFriends(friendsIds []social.ThirdPartyID, ID int64, tx *sql.Tx) error {
	queryString := fmt.Sprintf("INSERT IGNORE INTO `friends` (`user_id`, `friend_id`) "+
		"SELECT GREATEST(ids.id1, ids.id2),   LEAST(ids.id1, ids.id2) "+
		"FROM (SELECT ? as id1 , u2.id as id2 FROM (SELECT `id` FROM `users` u WHERE u.`%s` = ? ) u2) ids", friendsIds[0].Name())
	stmt, err := tx.Prepare(queryString)
	defer stmt.Close()
	for _, v := range friendsIds {
		_, err = stmt.Exec(ID, v.Value())
		if err != nil {
			return err
		}
	}
	return nil
}
