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

func (a *App) updateUserInfo(ctx iris.Context) {
	var newData map[string]interface{}
	user := ctx.Values().Get(userCtxKey).(*storage.User) //Вот тут явно криво, как бы вытащить данные методом getUser?
	userID := user.ID()
	oldData, err := user.Data()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Values().Set(errorCtxKey, err)
		return
	}
	err = ctx.ReadJSON(&newData)
	if err != nil {
		ctx.Values().Set("error", "updating user info error, read and parse failed. "+err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
		return
	}
	err = user.UpdateData(userID, oldData, newData)
	if err != nil {
		ctx.Values().Set("error", "updating user info db error. "+err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
	}
}
