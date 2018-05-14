package app

import (
	"gamelink-go/config"
	"gamelink-go/storage"
	"github.com/kataras/iris"
)

const (
	errorCtxKey = "error"
)

type (
	//Type to define json objects faster
	j map[string]interface{}

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
	users := a.iris.Party("/users", a.authMiddleware)
	{
		users.Get("/", a.getUser)
		users.Post("/", a.postUser)
		users.Delete("/", a.delete)
		users.Get("/addAuth", a.addAuth)
	}
	//service := i.Party("/service")
	//{
	//
	//}
	a.iris.OnAnyErrorCode(func(ctx iris.Context) {
		if config.IsDevelopmentEnv() {
			if err := ctx.Values().Get(errorCtxKey); err != nil {
				ctx.JSON(j{"error": err.(error).Error()})
			}
		}
	})
	return
}

//Run - This function will initialize router for the application and try to start listening clients
func (a *App) Run() error {
	return a.iris.Run(iris.Addr(config.ServerAddress))
}
