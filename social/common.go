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
		UserInfo() (ThirdPartyUser, error) //social id, name, friendsIds, error
		//IsDummy - return true if this is dummy token
		IsDummy() bool
	}

	//ThirdPartyUser - interface for user Data
	ThirdPartyUser interface {
		ID() ThirdPartyID
		//Name - returns user name
		Name() string
		//Bdate - return user birthday
		BirthDate() int64
		//Sex - return user UserSex
		Sex() string
		//Email - return user email
		Email() string
		//Country - return user country
		Country() string
		//Friends - return user friends
		Friends() []ThirdPartyID
	}

	commonInfo struct {
		FullName    string `json:"name"`
		Bdate       int64  `json:"bdate,omitempty"`
		UserSex     string `json:"sex,omitempty"`
		UserEmail   string `json:"email,omitempty"`
		UserCountry string `json:"country,omitempty"`
		friends     []ThirdPartyID
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

//Name - user name
func (d commonInfo) Name() string {
	return d.FullName
}

//Age - user age
func (d commonInfo) BirthDate() int64 {
	return d.Bdate
}

//Sex - user UserSex
func (d commonInfo) Sex() string {
	return d.UserSex
}

//Email - user email
func (d commonInfo) Email() string {
	return d.UserEmail
}

//Country - user email
func (d commonInfo) Country() string {
	return d.UserCountry
}

//Friends - return user friends
func (d commonInfo) Friends() []ThirdPartyID {
	return d.friends
}
