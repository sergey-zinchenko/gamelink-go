package social

import (
	"net/http"
	"time"
)

type (
	//ThirdPartyID - interface for
	ThirdPartyID interface {
		//Name - returns name of a identifier that can be used as database field name
		Name() string
		//Value - returns identifier value i.e. identifier it self.
		Value() string
	}

	//ThirdPartyToken - common interface for classes which can be used to obtain information of validity and user info of the third party tokens
	ThirdPartyToken interface {
		//UserInfo - get user info or error (d = NotFound if token is invalid or obsolete)
		UserInfo() (ThirdPartyID, string, []string, error) //social id, name, friendsIds, error
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
