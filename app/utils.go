package app

import (
	"gamelink-go/graceful"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"net/http"
)

func handleError(err error, ctx iris.Context) {
	if code, hasCode := err.(graceful.StatusCode); hasCode {
		ctx.StatusCode(code.StatusCode())
	} else {
		ctx.StatusCode(http.StatusInternalServerError)
	}
	logrus.Warn(errorCtxKey, err)
	ctx.Values().Set(errorCtxKey, err)
}
