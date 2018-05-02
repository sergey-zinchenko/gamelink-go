package app

import (
	"database/sql"
	"gamelink-go/config"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql" //That blank import is required to add mysql driver to the app
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
		Redis *redis.Client
		MySQL *sql.DB
	}
)

//NewApp - You can construct and initialize App (application) object with that function
//databases connections will be established and tested at the end of this function execution
func NewApp() (a *App, err error) {
	a = new(App)
	if a.MySQL, err = sql.Open("mysql", config.MysqlDsn); err != nil {
		a = nil
		return
	}
	a.MySQL.SetMaxIdleConns(10)
	if err = a.MySQL.Ping(); err != nil {
		a = nil
		return
	}
	a.Redis = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDb,
	})
	if _, err = a.Redis.Ping().Result(); err != nil {
		a = nil
	}
	return
}

//Run - This function will initialize router for the application and try to start listening clients
func (a *App) Run() error {
	i := iris.New()
	auth := i.Party("/auth")
	{
		auth.Get("/", a.registerLogin)
	}
	users := i.Party("/users", a.authMiddleware)
	{
		users.Get("/", a.getUser)
	}
	//service := i.Party("/service")
	//{
	//
	//}
	i.OnAnyErrorCode(func(ctx iris.Context) {
		if config.IsDevelopmentEnv() {
			if err := ctx.Values().Get(errorCtxKey); err != nil {
				ctx.JSON(j{"error": err.(error).Error()})
			}
		}
	})
	return i.Run(iris.Addr(config.ServerAddress))
}
