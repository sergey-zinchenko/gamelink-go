package social

import (
	"net/http"
	"time"
)

type (
	//ThirdPartyToken - common interface for classes which can be used to obtain information of validity and user info of the third party tokens
	ThirdPartyToken interface {
		//UserInfo - get user info or error (d = NotFound if token is invalid or obsolete)
		UserInfo() (string, string, error) //social id, name, error
	}
)

var client *http.Client

func init() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 100,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	client = &http.Client{Transport: tr}
}
