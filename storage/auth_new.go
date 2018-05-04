package storage

import (
	"database/sql"
	"errors"
	"gamelink-go/common"
	"gamelink-go/social"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

type (
	//tokenSource - type for enumeration of all possible sources (social networks) of the token for login & register procedure fo the system
	tokenSource int
)

const (
	//fbSource - mark given token as Facebook token
	fbSource tokenSource = iota
	//vkSource - mark given token as Vkontakte token
	vkSource
)

const (
	authRedisKeyPref = "auth:"
)

func check(source tokenSource, socialID string, tx *sql.Tx) (bool, int64, error) {
	var stmt *sql.Stmt
	var err error
	switch source {
	case vkSource:
		stmt, err = tx.Prepare("SELECT `id` FROM `users` u WHERE u.`vk_id` = ?")
	case fbSource:
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

func register(source tokenSource, socialID, name string, tx *sql.Tx) (int64, error) {
	var stmt *sql.Stmt
	var err error
	switch source {
	case vkSource:
		stmt, err = tx.Prepare("INSERT INTO `users` (`vk_id`, `name`) VALUES (?, ?)")
	case fbSource:
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

//AuthorizedUser - function to check our own authorization token from header. Returns user or nil if not valid token.
func (dbs DBS) AuthorizedUser(token string) (*User, error) {
	idStr, err := dbs.rc.Get(authRedisKeyPref + token).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
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
	var transaction = func(source tokenSource, socialID, name string, tx *sql.Tx) (int64, error) {
		registered, userID, err := check(source, socialID, tx)
		if err != nil {
			return 0, err
		}
		if !registered {
			if userID, err = register(source, socialID, name, tx); err != nil {
				return 0, err
			}
		}
		return userID, nil
	}
	var source tokenSource
	switch token.(type) {
	case social.VkToken:
		source = vkSource
	case social.FbToken:
		source = fbSource
	default:
		return nil, errors.New("unknown third party token type")
	}
	socialID, name, err := token.UserInfo()
	if err != nil {
		return nil, err
	}
	tx, err := dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	userID, err := transaction(source, socialID, name, tx)
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
	var authToken string
	for ok := false; !ok; {
		authToken = common.RandStringBytes(20)
		authKey := authRedisKeyPref + authToken
		var err error
		ok, err = u.dbs.rc.SetNX(authKey, u.ID(), time.Hour).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}
