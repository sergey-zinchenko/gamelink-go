package social

import (
	"time"
	"net/http"
	"gamelink-go/graceful"
)

var client *http.Client

func init() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client = &http.Client{Transport: tr}
}

type (
	TokenSource int
	IUserInfoGetter interface {
		GetUserInfo() (string, string, *graceful.Error)
	}
)

const (
	FbSource TokenSource = iota
	VKSource
)

func GetSocialUserInfo(source TokenSource, token string) (string, string, *graceful.Error) {
	var u IUserInfoGetter
	switch source {
	case FbSource:
		u = NewFbToken(token)
	case VKSource:
		u = NewVkToken(token)
	default:
		return "", "", graceful.NewInvalidError("invalid token source")
	}
	return u.GetUserInfo()
}
