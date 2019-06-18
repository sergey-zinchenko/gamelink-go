package app

import (
	C "gamelink-go/common"
	"gamelink-go/config"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/basicauth"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	errorCtxKey = "error"
)

type (
	//App structure - connects databases with the middleware and handlers of router
	App struct {
		dbs   *storage.DBS
		iris  *iris.Application
		ranks *storage.Ranks
	}
)

//ConnectDataBases - tries to connect to all databases required to function of the app. Method can be recalled.
func (a *App) ConnectDataBases() error {
	if err := a.dbs.Connect(); err != nil {
		return err
	}
	if err := a.dbs.CheckTables(); err != nil {
		return err
	}
	return nil
}

//GenerateRanks - invoke storage func that generate rank arrays
func (a *App) GenerateRanks(count int) {
	for {
		err := a.ranks.GenerateRankArrays(count)
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(config.UpdateLbArraysDataInSecondsPeriod)
	}
}

//NewApp - You can construct and initialize App (application) object with that function
//router will be configured but not database connections
func NewApp() (a *App) {
	a = new(App)
	a.iris = iris.New()
	a.dbs = &storage.DBS{}
	a.ranks = &storage.Ranks{DBS: a.dbs}

	fakedata := a.iris.Party("/fake")
	{
		fakedata.Get("/users", a.addFakeUsers)
		fakedata.Get("/token/{id:int}", a.addFakeToken)
	}

	auth := a.iris.Party("/auth")
	{
		auth.Get("/", a.registerLogin)
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
		instances.Get("/{id:int}", a.getSave)
		instances.Post("/", a.postSave)
		instances.Post("/{id:int}", a.postSave)
		instances.Delete("/{id:int}", a.deleteSave)
	}
	leaderboards := a.iris.Party("/leaderboards", a.authMiddleware)
	{
		leaderboards.Get("/{id:int}/{lbtype: string}", a.getLeaderboard)
	}

	if config.TournamentsSupported {
		authConfig := basicauth.Config{
			Users:   map[string]string{config.TournamentsAdminUsername: config.TournamentsAdminPassword},
			Realm:   "Authorization Required", // defaults to "Authorization Required"
			Expires: time.Duration(30) * time.Minute,
		}

		authentication := basicauth.New(authConfig)

		needAuth := a.iris.Party("/tournaments", authentication)
		{
			needAuth.Get("/start", a.startTournament)
		}

		tournaments := a.iris.Party("/tournaments", a.authMiddleware)
		{
			tournaments.Get("/{tournament_id:int}/join", a.joinTournament)
			tournaments.Post("/{tournament_id:int}/updatescore", a.updateScore)
			tournaments.Get("/{tournament_id:int}/leaderboard", a.getRoomLeaderboard)
			tournaments.Get("/list", a.getAvailableTournaments)
			tournaments.Get("/results", a.getUsersResults)
			tournaments.Get("/next", a.timeToNextTournament)
		}
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
