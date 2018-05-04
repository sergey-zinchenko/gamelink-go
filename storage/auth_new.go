package storage

import (
	"github.com/go-redis/redis"
	"strconv"
)

//AuthorizedUser - function to check our own authorization token from header. Returns user or nil if not valid token.
func (dbs DBS) AuthorizedUser(authToken string) (*User, error) {
	idStr, err := dbs.rc.Get(authRedisKeyPref + authToken).Result()
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
