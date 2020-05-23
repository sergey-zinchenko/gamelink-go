package storage

import (
	"context"
	"errors"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	//AuthRedisKeyPref - auth redis key prefix
	AuthRedisKeyPref = "auth:"
)

//AuthorizedUser - function to check our own authorization token from header. Returns user or nil if not valid token.
func (dbs DBS) AuthorizedUser(ctx context.Context, token string) (*User, error) {
	if dbs.rc == nil {
		return nil, errors.New("databases not initialized")
	}
	if len(token) < 6 {
		return nil, graceful.UnauthorizedError{Message: "token too short"}
	}
	var id int64
	isDummy := token[:5] == "dummy"
	tokenWithPrefix := AuthRedisKeyPref + token
	for {
		err := dbs.rc.Watch(ctx, func(tx *redis.Tx) error {
			var idStr string
			var err error
			idStr, err = tx.Get(ctx, tokenWithPrefix).Result()
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
			_, err = tx.Set(ctx, tokenWithPrefix, id, lifetime).Result()
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
func (dbs DBS) AuthToken(ctx context.Context, generateDummyToken bool, id int64) (string, error) {
	if dbs.rc == nil {
		return "", errors.New("databases not initialized")
	}
	var authToken, authKey string
	for ok := false; !ok; {
		var err error
		var lifetime time.Duration
		if !generateDummyToken {
			authToken = C.RandStringBytes(40)
			for authToken[:5] == "dummy" {
				authToken = C.RandStringBytes(40)
			}
			authKey = AuthRedisKeyPref + authToken
			lifetime = time.Hour
		} else {
			authToken = "dummy" + C.RandStringBytes(35)
			authKey = AuthRedisKeyPref + authToken
			lifetime = 24 * 30 * time.Hour
		}
		ok, err = dbs.rc.SetNX(ctx, authKey, id, lifetime).Result()
		if err != nil {
			return "", err
		}
	}
	return authToken, nil
}

//DeleteRedisToken - delete token from redis
func (dbs DBS) DeleteRedisToken(ctx context.Context, token string) error {
	cmd := dbs.rc.Del(ctx, AuthRedisKeyPref+token)
	if cmd.Err() != nil {
		logrus.Warn("redis delete token error", cmd.Err())
		return cmd.Err()
	}
	return nil
}
