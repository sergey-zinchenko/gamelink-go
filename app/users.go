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
	user := ctx.Values().Get(userCtxKey).(*storage.User)
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

func (a *App) deleteUserInfo(ctx iris.Context) {
	var err error
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	userID := user.ID()
	Data, err := user.Data()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Values().Set(errorCtxKey, err)
		return
	}
	queryValues := ctx.Request().URL.Query()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Values().Set(errorCtxKey, err)
		return
	}
	err = user.DeleteData(userID, queryValues, Data)
	if err != nil {
		ctx.Values().Set("error", "deleting user info db error. "+err.Error())
		ctx.StatusCode(iris.StatusInternalServerError)
	}
}

func (a *App) addAnotherSocialAcc(ctx iris.Context) {
	var err error
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	userID := user.ID()
	queryValues := ctx.Request().URL.Query()
	if err != nil {
		ctx.Values().Set("error", "bad request ."+err.Error())
		return
	}
	err = user.AddSocialAcc(userID, queryValues)
	if err != nil {
		//ctx.Values().Set("error", "adding social account error. ")
		ctx.StatusCode(iris.StatusInternalServerError)
	}
}
