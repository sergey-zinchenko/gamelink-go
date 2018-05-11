package app

import (
	"gamelink-go/graceful"
	"github.com/kataras/iris"
	"net/http"
)

func handleError(err error, ctx iris.Context) {
	if err == nil {
		return
	}
	if code, hasCode := err.(graceful.StatusCode); hasCode {
		ctx.StatusCode(code.StatusCode())
	} else {
		ctx.StatusCode(http.StatusInternalServerError)
	}
	ctx.Values().Set(errorCtxKey, err)
	ctx.StopExecution()
}
