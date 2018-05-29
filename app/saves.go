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
	saveID, _ := ctx.Params().GetInt("id")
	instances, err := user.Saves(saveID)
	if err != nil {
		handleError(err, ctx)
		return
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
	saveID, _ := ctx.Params().GetInt("id")
	if saveID != 0 {
		save, err = user.UpdateSave(data, saveID)
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
	saveID, _ := ctx.Params().GetInt("id")
	data, err = user.DeleteSave(saveID, ctx.Request().URL.Query()["fields"])
	if err != nil {
		return
	}
	if data == nil {
		ctx.StatusCode(http.StatusNoContent)
	} else {
		ctx.JSON(data)
	}

}
