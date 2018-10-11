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
	authRedisKeyPref  = "auth:"
	dummyRedisKeyPref = "dummy:"
)

//AuthorizedUser - function to check our own authorization token from header. Returns user or nil if not valid token.
func (dbs DBS) AuthorizedUser(token string) (*User, error) {
	if dbs.rc == nil {
		return nil, errors.New("databases not initialized")
	}
	var id int64
	err := dbs.rc.Watch(func(tx *redis.Tx) error {
		var isDummy bool
		var idStr string
		var err error
		idStr, err = dbs.rc.Get(authRedisKeyPref + token).Result()
		if err != nil && err != redis.Nil {
			return err
		} else if err == redis.Nil {
			idStr, err = dbs.rc.Get(dummyRedisKeyPref + token).Result()
			if err != nil {
				if err == redis.Nil {
					return graceful.UnauthorizedError{Message: "key doesn't exist in redis"}
				}
				return err
			}
			isDummy = true
		}
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return err
		}
		if !isDummy {
			_, err = dbs.rc.Set(authRedisKeyPref+token, id, 8*time.Hour).Result()
		} else {
			_, err = dbs.rc.Set(dummyRedisKeyPref+token, id, 24*30*time.Hour).Result()
		}
		return err
	}, authRedisKeyPref+token)
	if err != nil {
		return nil, err
	}
	return &User{id, &dbs}, nil
}

//ThirdPartyUser - function to login or register user using his third party token
func (dbs DBS) ThirdPartyUser(token social.ThirdPartyToken, deviceID string, deviceType string) (*User, error) {
	var device *Device
	var err error
	if dbs.mySQL == nil {
		return nil, errors.New("databases not initialized")
	}
	u := User{dbs: &dbs}
	if deviceID != "" {
		device = &Device{deviceID: deviceID, deviceType: deviceType}
	}
	if token != nil {
		err = u.LoginUsingThirdPartyToken(token, device)
	} else {
		err = u.LoginDummy(device)
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

//AuthToken - Function to generate and store auth token in rc.
func (u User) AuthToken() (string, error) {
	if u.dbs.rc == nil {
		return "", errors.New("databases not initialized")
	}
	var authToken string
	for ok := false; !ok; {
		authToken = C.RandStringBytes(40)
		authKey := authRedisKeyPref + authToken
		var err error
		ok, err = u.dbs.rc.SetNX(authKey, u.ID(), time.Hour).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//DummyToken - Function to generate and store dummy token in rc.
func (u User) DummyToken() (string, error) {
	if u.dbs.rc == nil {
		return "", errors.New("databases not initialized")
	}
	var authToken string
	for ok := false; !ok; {
		authToken = C.RandStringBytes(40)
		authKey := dummyRedisKeyPref + authToken
		var err error
		ok, err = u.dbs.rc.SetNX(authKey, u.ID(), 24*30*time.Hour).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//AddDeviceID - add deviceID to db
func (u User) AddDeviceID(deviceID string, deviceType string) error {
	_, err := u.dbs.mySQL.Exec(queries.AddDeviceID, u.ID(), deviceID, deviceType)
	if err != nil {
		return err
	}
	return nil
}
