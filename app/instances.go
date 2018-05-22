package app

import (
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

func (a *App) getAllUserInstances(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	instances, err := user.Instances(ctx.Request().URL.Query()["id"])
	if err != nil {
		handleError(err, ctx)
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.Text(instances)
}
