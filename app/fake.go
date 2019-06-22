package app

import (
	"fmt"
	C "gamelink-go/common"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	fakeCount = 250000
)

func (a *App) addFakeUsers(ctx iris.Context) {
	t0 := time.Now()
	fmt.Println(fmt.Sprintf("creating %d users... Started at : %v", fakeCount, t0))
	a.dbs.AddFakeUser(fakeCount)
	fmt.Println(fmt.Sprintf(" __________users addition ended...let's make a friendship!!!!!!!!!!!!! elapsed: %v", time.Since(t0)))
	a.dbs.AddFakeFriends(fakeCount)
	fmt.Println(fmt.Sprintf("succesfully added %d users; elapsed: %v", fakeCount, time.Since(t0)))
	ctx.StatusCode(http.StatusOK)
}

func (a *App) addFakeToken(ctx iris.Context) {
	var userID int
	var err error
	if ctx.Params().GetEntry("id").ValueRaw != nil {
		userID, err = ctx.Params().GetInt("id")
		if err != nil {
			return
		}
	}
	authToken, err := a.dbs.AuthToken(true, int64(userID))
	if err != nil {
		logrus.Warn(err.Error())
		return
	}
	ctx.JSON(C.J{"token": authToken})
}
