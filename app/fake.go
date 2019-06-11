package app

import (
	"fmt"
	"github.com/kataras/iris"
	"net/http"
	"time"
)

const (
	fakeCount = 10000
)

func (a *App) addFakeUsers(ctx iris.Context) {
	t0 := time.Now()
	fmt.Println(fmt.Sprintf("started at : %v", t0))
	a.dbs.AddFakeUser(fakeCount)
	fmt.Println(fmt.Sprintf(" __________users addition ended...let's make a friendship!!!!!!!!!!!!! elapsed: %v", time.Since(t0)))
	a.dbs.AddFakeFriends(fakeCount)
	fmt.Println(fmt.Sprintf("succesfully added %d users; elapsed: %v", fakeCount, time.Since(t0)))
	ctx.StatusCode(http.StatusOK)
}
