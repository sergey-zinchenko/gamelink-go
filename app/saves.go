package app

import (
	C "gamelink-go/common"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"net/http"
	"time"
)

func (a *App) getSave(ctx iris.Context) {
	var saveID int
	var err error
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	if ctx.Params().GetEntry("id").ValueRaw != nil {
		saveID, err = ctx.Params().GetInt("id")
		if err != nil {
			return
		}
	}
	instances, err := user.SavesString(saveID)
	if err != nil {
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(instances)
}

func (a *App) postSave(ctx iris.Context) {
	var (
		data, save C.J
		saveID     int
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
	_, flag := ctx.Params().GetEntry("id")
	if flag != false {
		saveID, err = ctx.Params().GetInt("id")
		if err != nil {
			return
		}
	}
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
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	saveID, err := ctx.Params().GetInt("id")
	if err != nil {
		return
	}
	data, err = user.DeleteSave(saveID, ctx.Request().URL.Query()["fields"])
	if err != nil {
		return
	}
	if data == nil {
		ctx.StatusCode(http.StatusNoContent)
	} else {
		data["updated_at"] = time.Now().Unix()
		ctx.JSON(data)
	}
}
