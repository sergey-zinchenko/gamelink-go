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
		iris  *iris.Application
	}
)

//ConnectDataBases - tries to connect to all databases required to function of the app. Method can be recalled.
func (a *App) ConnectDataBases() error {
	var err error
	if a.MySQL == nil {
		if a.MySQL, err = sql.Open("mysql", config.MysqlDsn); err != nil {
			return err
		}
		a.MySQL.SetMaxIdleConns(10)
		if err = a.MySQL.Ping(); err != nil {
			a.MySQL.Close() //TODO: нужно проверить правильная ли это вообще мысль
			a.MySQL = nil
			return err
		}
	}
	if a.Redis == nil {
		a.Redis = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDb,
		})
		if _, err = a.Redis.Ping().Result(); err != nil {
			a.Redis.Close() //TODO: нужно проверить правильная ли это вообще мысль
			a.Redis = nil
			return err
		}
	}
	return nil
}

//NewApp - You can construct and initialize App (application) object with that function
//router will be configured but not database connections
func NewApp() (a *App) {
	a = new(App)
	a.iris = iris.New()
	auth := a.iris.Party("/auth")
	{
		auth.Get("/", a.registerLogin)
	}
	users := a.iris.Party("/users", a.authMiddleware)
	{
		users.Get("/", a.getUser)
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
