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

func (a *App) postUser(ctx iris.Context) {
	var (
		data, updated map[string]interface{}
		err           error
	)
	defer func() {
		handleError(err, ctx)
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	err = ctx.ReadJSON(&data)
	if err != nil {
		return
	}
	updated, err = user.Update(data)
	if err != nil {
		return
	}
	ctx.JSON(updated)
}

func (a *App) delete(ctx iris.Context) {
	var (
		err  error
		data map[string]interface{}
	)
	defer func() {
		handleError(err, ctx)
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err = user.Delete(ctx.Request().URL.Query()["fields"])
	if err != nil {
		return
	}
	if data == nil {
		ctx.StatusCode(http.StatusNoContent)
	} else {
		ctx.JSON(data)
	}
}

func (a *App) addSocial(ctx iris.Context) {
	var (
		err  error
		data map[string]interface{}
	)
	defer func() {
		handleError(err, ctx)
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err = user.AddSocial(ctx.Request().URL.Query())
	if err != nil {
		return
	}
	ctx.JSON(data)
}
