package app

import (
	"github.com/kataras/iris"
)

func (a *App) getUser(ctx iris.Context) {
	userId := ctx.Values().Get(userIdValueKey).(int64)
	ctx.JSON(J{"userId": userId})
}