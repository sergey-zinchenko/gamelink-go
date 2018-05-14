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
	queryString := fmt.Sprintf("SELECT `id` FROM `users` u WHERE u.`%s` = ?", socialID.Name())
	stmt, err := tx.Prepare(queryString) //TODO: do not use Prepare and close in one func
	if err != nil {
		return false, 0, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(socialID) //TODO: QueryRow + sql.ErrNoRows
	if err != nil {
		return false, 0, err
	}
	defer rows.Close()
	registered := rows.Next()
	//TODO: rows.Err()
	var userID int64
	if registered {
		err = rows.Scan(&userID)
		if err != nil {
			return true, 0, err
		}
	}
	return registered, userID, nil
}

func register(socialID social.ThirdPartyID, name string, tx *sql.Tx) (int64, error) {
	stmt, err := tx.Prepare("INSERT INTO `users` (`data`) VALUES (?)")
	if err != nil {
		return 0, err
	}
	b, err := json.Marshal(map[string]interface{}{socialID.Name(): socialID, "name": name})
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(b)
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
	var transaction = func(socialID social.ThirdPartyID, name string, tx *sql.Tx) (int64, error) {
		registered, userID, err := check(socialID, tx)
		if err != nil {
			return 0, err
		}
		if !registered {
			if userID, err = register(socialID, name, tx); err != nil {
				return 0, err
			}
		}
		return userID, nil
	}
	socialID, name, friendsIds, err := token.UserInfo()
	if err != nil {
		return nil, err
	}
	// Вопрос в том, когда завершается горутина? Если при завершении функции, в которой вызвана, не все данные попадуn в таблицу.
	// Идея была сделать процес фоном и быстрее залогинить пользователя чтоб он не ждал пока база запишет друзей
	if friendsIds != nil {
		go dbs.SyncFriends(friendsIds, socialID.Value())
	}
	tx, err := dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	userID, err := transaction(socialID, name, tx)
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
func (dbs DBS) SyncFriends(friendsIds []string, ID string) error {
	var transaction = func(friendsIds []string, ID string, tx *sql.Tx) error {
		for _, v := range friendsIds {
			//err := tx.QueryRow("SELECT `user_id` FROM `friends` f WHERE f.`user_id`=? AND f.`friend_id`=?", ID, v) // Вызывает ошибку busy buffer
			//if err != nil {
			stmt, err := tx.Prepare("INSERT INTO `friends` (`user_id`, `friend_id`) VALUES (?,?)")
			if err != nil {
				return err
			}
			defer stmt.Close()
			_, err = stmt.Exec(ID, v)
			if err != nil {
				return err
			}
			//}
		}
		return nil
	}
	tx, err := dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(friendsIds, ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
