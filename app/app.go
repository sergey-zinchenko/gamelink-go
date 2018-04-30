package app

import (
	"github.com/kataras/iris"
	"gamelink-go/config"
	"github.com/go-redis/redis"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	errorCtxKey = "error"
)

type App struct {
	Redis *redis.Client
	MySql *sql.DB
}

func NewApp() (a *App, err error) {
	a = new(App)
	if a.MySql, err = sql.Open("mysql", config.MysqlDsn); err != nil {
		a = nil
		return
	}
	a.MySql.SetMaxIdleConns(10)
	if err = a.MySql.Ping(); err != nil {
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

func (a *App) Run() error {
	i := iris.New()
	auth := i.Party("/auth")
	{
		auth.Get("/", a.registerLogin2)
	}
	users := i.Party("users", a.authMiddleware2)
	{
		users.Get("/", a.getUser)
	}
	i.OnAnyErrorCode(func(ctx iris.Context) {
		if config.IsDevelopmentEnv() {
			if err := ctx.Values().Get(errorCtxKey); err != nil {
				ctx.JSON(J{"error": err.(error).Error()})
			}
		}
	})
	return i.Run(iris.Addr(config.ServerAddress))
}