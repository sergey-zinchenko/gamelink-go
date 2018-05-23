package app

import (
	C "gamelink-go/common"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"net/http"
)

func (a *App) getSave(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	instances, err := user.Saves(ctx.Request().URL.Query()["id"])
	if err != nil {
		handleError(err, ctx)
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(instances)
}

func (a *App) postSave(ctx iris.Context) {
	var (
		data, save C.J
		err        error
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	err = ctx.ReadJSON(&data)
	if err != nil {
		return
	}
	saveID := ctx.Request().URL.Query()["id"]
	if len(saveID) != 0 {
		save, err = user.UpdateSave(data, saveID[0])
	} else {
		save, err = user.CreateSave(data)
	}
	if err != nil {
		return
	}
	ctx.JSON(save)
}

func (a *App) deleteSave(ctx iris.Context) {
	var (
		data C.J
		err  error
	)
	defer func() {
		handleError(err, ctx)
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err = user.DeleteSave(ctx.Request().URL.Query()["id"], ctx.Request().URL.Query()["fields"])
	if err != nil {
		return
	}
	if data == nil {
		ctx.StatusCode(http.StatusNoContent)
	} else {
		ctx.JSON(data)
	}
}
