package social

import (
	"gamelink-go/graceful"
	"github.com/dustinkirkland/golang-petname"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type (
	//DummyToken - Class to get information about Facebook user tokens
	DummyToken string

	dummyError struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	//DummyIdentifier - class to store fb identifier and column name
	DummyIdentifier string
	//DummyInfo - class to user info from FB
	DummyInfo struct {
		DummyID DummyIdentifier `json:"dummy_id"`
		commonInfo
	}
)

const (
	//DummyID - const name of facebook identifier column in the db
	DummyID = "dummy_id"
)

//Name - column name in the db
func (i DummyIdentifier) Name() string {
	return DummyID
}

//Value - dummy id value
func (i DummyIdentifier) Value() string {
	return string(i)
}

//ID - return fbID
func (d DummyInfo) ID() ThirdPartyID {
	return d.DummyID
}

//IsDummy - return true if this user auth without social
func (d DummyInfo) IsDummy() bool {
	return true
}

func (token DummyToken) get(userInfo *DummyInfo) error {
	userInfo.FullName = petname.Generate(2, " ")
	return nil
}

//UserInfo - method to get user information (name and identifier) of a dummy user token
func (token DummyToken) UserInfo() (ThirdPartyUser, error) {
	if token == "" {
		return nil, graceful.UnauthorizedError{Message: "empty token"}
	}
	userInfo := DummyInfo{DummyIdentifier("1"), commonInfo{}}
	err := token.get(&userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}
