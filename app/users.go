package app

import (
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
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
	ctx.ContentType(context.ContentJSONHeaderValue)
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

func (a *App) addAuth(ctx iris.Context) {
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
	token := tokenFromValues(ctx.Request().URL.Query())
	if token == nil {
		err = graceful.BadRequestError{Message: "invalid token"}
		return
	}
	data, err = user.AddSocial(token)
	if err != nil {
		return
	}
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	arr := strings.Split(header, " ")
	if arr[1] != "" {
		err = user.SwitchRedisToken(arr[1])
		if err != nil {
			return
		}
	}
	ctx.JSON(data)
}
