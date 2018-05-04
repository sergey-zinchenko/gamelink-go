package app

import (
	"gamelink-go/storage"
	"github.com/kataras/iris"
)

func (a *App) getUser(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	ctx.JSON(j{"userID": user.ID()})
}
