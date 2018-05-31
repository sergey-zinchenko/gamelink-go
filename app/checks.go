package app

import (
	"github.com/kataras/iris"
	"log"
	"net/http"
	"sync/atomic"
)

func (a *App) healthCheck(ctx iris.Context) {
	ctx.StatusCode(http.StatusOK)
}

func (a *App) readyCheck(ctx iris.Context) {
	isReady := &atomic.Value{}
	isReady.Store(true)
	if isReady == nil || !isReady.Load().(bool) {
		log.Printf("ready check fail")
		ctx.StatusCode(http.StatusServiceUnavailable)
		return
	}
	ctx.StatusCode(http.StatusOK)
}
