package app

import (
	"github.com/kataras/iris"
)

func (a *App) getUser(ctx iris.Context) {
	userID := ctx.Values().Get(userCtxKey).(int64)
	ctx.JSON(j{"userID": userID})
}
