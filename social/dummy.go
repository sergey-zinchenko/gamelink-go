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
	DummyToken struct {
		info DummyInfo
	}

	//DummyInfo - class to dummy user info
	DummyInfo struct {
		commonInfo
	}
)

//NewDummyToken - create new dummy token struct with random generated name
func NewDummyToken() DummyToken {
	d := DummyToken{info: DummyInfo{}}
	d.info.FullName = petname.Generate(2, " ")
	return d
}

//ID - return dummyID
func (d DummyInfo) ID() ThirdPartyID {
	return nil
}

//UserInfo - return dummy user info with generated name
func (d DummyToken) UserInfo() (ThirdPartyUser, error) {
	return d.info, nil
}

//IsDummy - true cause it's dummy token
func (d DummyToken) IsDummy() bool {
	return true
}
