package app

import "C"
import (
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
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
