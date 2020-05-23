package app

import (
	"context"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	irisConext "github.com/kataras/iris/context"
	"net/http"
	"strings"
)

func (a *App) getUser(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err := user.DataString()
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(irisConext.ContentJSONHeaderValue)
	ctx.WriteString(data)
}

func (a *App) postUser(ctx iris.Context) {
	var (
		data, updated C.J
		err           error
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
	updated, err = user.Update(data)
	if err != nil {
		return
	}
	ctx.JSON(updated)
}

func (a *App) deleteUser(ctx iris.Context) {
	var (
		err  error
		data C.J
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
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

func (a *App) addAuth(ctx irisConext.Context) {
	var (
		err       error
		data      C.J
		existedID int64
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	token := tokenFromValues(ctx.Request().URL.Query())
	if token == nil {
		err = graceful.BadRequestError{Message: "invalid token"}
		return
	}
	data, existedID, err = user.AddSocial(token)
	if err != nil {
		return
	}
	response := make(map[string]interface{})
	response["data"] = data
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	arr := strings.Split(header, " ")
	tokenValue := arr[1]
	if tokenValue != "" && tokenValue[:5] == "dummy" {
		var updID int64
		if existedID != 0 {
			updID = existedID
		} else {
			updID = user.ID()
		}
		newToken, err := a.dbs.AuthToken(context.Background(), false, updID) //false cause we want generate new token for normal user not for dummy
		if err != nil {
			return
		}
		err = a.dbs.DeleteRedisToken(context.Background(), tokenValue)
		if err != nil {
			return
		}
		response["token"] = newToken
	}
	_, err = ctx.JSON(response)
}
