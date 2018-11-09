package social

import (
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

	//DummyIdentifier - class to store fb identifier and column name
	DummyIdentifier string
	//DummyInfo - class to dummy user info
	DummyInfo struct {
		//DummyID DummyIdentifier `json:"dummy_id"`
		commonInfo
	}
)

const (
	//DummyID - const name of Dummy identifier column in the db
	DummyID = ""
)

//Name - dummy column name in the db
func (d DummyIdentifier) Name() string {
	return DummyID
}

//Value - fb id value
func (d DummyIdentifier) Value() string {
	return string(d)
}

//ID - return dummyID
func (d DummyInfo) ID() ThirdPartyID {
	return nil
}

func (token DummyToken) get(userInfo *DummyInfo) error {
	userInfo.FullName = petname.Generate(2, " ")
	return nil
}

//UserInfo - return dummy user info with generated name
func (token DummyToken) UserInfo() (ThirdPartyUser, error) {
	userInfo := DummyInfo{commonInfo{}}
	err := token.get(&userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}
