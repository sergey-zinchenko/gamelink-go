package app

import (
	stdContext "context"
	C "gamelink-go/common"
	"gamelink-go/config"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"time"
)

const (
	errorCtxKey = "error"
)

type (
	//App structure - connects databases with the middleware and handlers of router
	App struct {
		dbs  *storage.DBS
		iris *iris.Application
	}
)

//ConnectDataBases - tries to connect to all databases required to function of the app. Method can be recalled.
func (a *App) ConnectDataBases() error {
	return a.dbs.Connect()
}

//NewApp - You can construct and initialize App (application) object with that function
//router will be configured but not database connections
func NewApp() (a *App) {
	a = new(App)
	a.iris = iris.New()
	a.dbs = &storage.DBS{}
	auth := a.iris.Party("/auth")
	{
		auth.Get("/", a.registerLogin)
	}
	service := a.iris.Party("/service")
	{

		service.Get("/", a.getAppInfo)
		service.Get("/healthz", a.healthCheck)
		service.Get("/readyz", a.readyCheck)
	}
	users := a.iris.Party("/users", a.authMiddleware)
	{
		users.Get("/", a.getUser)
		users.Post("/", a.postUser)
		users.Delete("/", a.deleteUser)
		users.Get("/addAuth", a.addAuth)
	}
	instances := a.iris.Party("/saves", a.authMiddleware)
	{
		instances.Get("/", a.getSave)
		instances.Get("/{id}", a.getSave)
		instances.Post("/", a.postSave)
		instances.Post("/{id}", a.postSave)
		instances.Delete("/{id}", a.deleteSave)
	}
	leaderboards := a.iris.Party("/leaderboards", a.authMiddleware)
	{
		leaderboards.Get("/{id:int}/{lbtype: string}", a.getLeaderboard)
	}
	a.iris.OnAnyErrorCode(func(ctx iris.Context) {
		if config.IsDevelopmentEnv() {
			if err := ctx.Values().Get(errorCtxKey); err != nil {
				ctx.JSON(C.J{"error": err.(error).Error()})
			}
		}
	})
	return
}

//Run - This function will initialize router for the application and try to start listening clients
func (a *App) Run() error {
	return a.iris.Run(iris.Addr(config.ServerAddress))
}

//Shutdown - make Gracefull Shutdown
func (a *App) Shutdown() {
	iris.RegisterOnInterrupt(func() {
		timeout := 5 * time.Second
		ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
		defer cancel()
		// close all hosts
		a.iris.Shutdown(ctx)
	})
}
