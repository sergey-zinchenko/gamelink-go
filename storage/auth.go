package storage

import (
	"errors"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"gamelink-go/storage/queries"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

const (
	authRedisKeyPref = "auth:"
)

//AuthorizedUser - function to check our own authorization token from header. Returns user or nil if not valid token.
func (dbs DBS) AuthorizedUser(token string) (*User, error) {
	if dbs.rc == nil {
		return nil, errors.New("databases not initialized")
	}
	var id int64
	isDummy := token[:5] == "dummy"
	tokenWithPrefix := authRedisKeyPref + token
	for {
		err := dbs.rc.Watch(func(tx *redis.Tx) error {
			var idStr string
			var err error
			idStr, err = tx.Get(tokenWithPrefix).Result()
			if err != nil {
				if err == redis.Nil {
					return graceful.UnauthorizedError{Message: "key doesn't exist in redis"}
				}
				return err
			}
			id, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return err
			}
			var lifetime time.Duration
			if isDummy {
				lifetime = 24 * 30 * time.Hour
			} else {
				lifetime = 8 * time.Hour
			}
			_, err = tx.Set(tokenWithPrefix, id, lifetime).Result()
			return err
		}, tokenWithPrefix)
		if err != nil {
			if err == redis.TxFailedErr {
				continue
			} else {
				return nil, err
			}
		}
		break
	}
	return &User{id, &dbs}, nil
}

//ThirdPartyUser - function to login or register user using his third party token
func (dbs DBS) ThirdPartyUser(token social.ThirdPartyToken) (*User, error) {
	var err error
	if dbs.mySQL == nil {
		return nil, errors.New("databases not initialized")
	}
	u := User{dbs: &dbs}
	err = u.LoginUsingThirdPartyToken(token)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

//AuthToken - Function to generate and store auth token in rc.
func (u User) AuthToken(isDummy bool) (string, error) {
	if u.dbs.rc == nil {
		return "", errors.New("databases not initialized")
	}
	var authToken, authKey string
	for ok := false; !ok; {
		var err error
		var lifetime time.Duration
		if !isDummy {
			authToken = C.RandStringBytes(40)
			for authToken[:5] == "dummy" {
				authToken = C.RandStringBytes(40)
			}
			authKey = authRedisKeyPref + authToken
			lifetime = time.Hour
		} else {
			authToken = "dummy" + C.RandStringBytes(35)
			authKey = authRedisKeyPref + authToken
			lifetime = 24 * 30 * time.Hour
		}
		ok, err = u.dbs.rc.SetNX(authKey, u.ID(), lifetime).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//AddDeviceID - add deviceID to db
func (u User) AddDeviceID(deviceID string, deviceType string) error {
	if deviceID != "" && deviceType != "" {
		_, err := u.dbs.mySQL.Exec(queries.AddDeviceID, u.ID(), deviceID, deviceType)
		if err != nil {
			return err
		}
	}
	return nil
}
