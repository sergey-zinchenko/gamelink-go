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
	var isDummy bool
	if token[:5] == "dummy" {
		isDummy = true
	}
	//TODO: надо вынести authRedisKeyPref + token за цикл и вообще в отдельную переменнную
	for {
		err := dbs.rc.Watch(func(tx *redis.Tx) error {
			var idStr string
			var err error
			idStr, err = tx.Get(authRedisKeyPref + token).Result()
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
			//TODO: плохой стиль в условиях вызвывать одну и туже функцию - по условию меняй переменную а потом вызывай функцию с ней в качестве параметра
			if !isDummy {
				_, err = tx.Set(authRedisKeyPref+token, id, 8*time.Hour).Result()
			} else {
				_, err = tx.Set(authRedisKeyPref+token, id, 24*30*time.Hour).Result()
			}
			return err
		}, authRedisKeyPref+token)
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
		if isDummy == false {
			authToken = C.RandStringBytes(40)
			if authToken[:5] == "dummy" {
				authToken = C.RandStringBytes(40)
			}
			authKey = authRedisKeyPref + authToken
			ok, err = u.dbs.rc.SetNX(authKey, u.ID(), time.Hour).Result()
		} else {
			authToken = "dummy" + C.RandStringBytes(35)
			authKey = authRedisKeyPref + authToken
			ok, err = u.dbs.rc.SetNX(authKey, u.ID(), 24*30*time.Hour).Result() //TODO: та же история с этой функцией и временем - вызывай ее за условным оператором одни раз
		}
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//AddDeviceID - add deviceID to db
func (u User) AddDeviceID(deviceID string, deviceType string) error {
	//TODO: нужно модифицировать если применять правки из апп/аус
	_, err := u.dbs.mySQL.Exec(queries.AddDeviceID, u.ID(), deviceID, deviceType)
	if err != nil {
		return err
	}
	return nil
}
