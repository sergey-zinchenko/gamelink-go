package storage

import (
	"errors"
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
	u := User{dbs: &dbs}
	if err := u.LoginUsingThirdPartyToken(token); err != nil {
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
