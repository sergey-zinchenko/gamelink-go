package app

import (
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"net/http"
)

func (a *App) getUser(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err := user.Data()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Values().Set(errorCtxKey, err)
		return
	}
	ctx.JSON(data)
}
