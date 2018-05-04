package social

import (
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
	//IUserInfoGetter - common interface for classes which can be used to obtain information of validity and user info of the third party tokens
	IUserInfoGetter interface {
		//GetUserInfo - get user info or error (d = NotFound if token is invalid or obsolete)
		GetUserInfo() (string, string, error)
	}
)
