package social

import (
	"errors"
	"net/http"
	"time"
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
	//TokenSource - type for enumeration of all possible sources (social networks) of the token for login & register procedure fo the system
	TokenSource int
	//IUserInfoGetter - common interface for classes which can be used to obtain information of validity and user info of the third party tokens
	IUserInfoGetter interface {
		//GetUserInfo - get user info or error (d = NotFound if token is invalid or obsolete)
		GetUserInfo() (string, string, error)
	}
)

const (
	//FbSource - mark given token as Facebook token
	FbSource TokenSource = iota
	//VKSource - mark given token as Vkontakte token
	VKSource
)

//GetSocialUserInfo - common function to get information about given token from source
func GetSocialUserInfo(source TokenSource, token string) (string, string, error) {
	var u IUserInfoGetter
	switch source {
	case FbSource:
		u = NewFbToken(token)
	case VKSource:
		u = NewVkToken(token)
	default:
		return "", "", errors.New("invalid token source")
	}
	return u.GetUserInfo()
}
