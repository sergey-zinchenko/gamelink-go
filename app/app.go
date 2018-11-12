package app

import (
	"gamelink-go/admingrpc"
	"gamelink-go/adminnats"
	C "gamelink-go/common"
	"gamelink-go/config"
	service "gamelink-go/proto_service"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/basicauth"
	"github.com/nats-io/go-nats"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
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
		admin *admingrpc.AdminServiceServer
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

//ConnetcGRPC - tries to make grpc connection
func (a *App) ConnetcGRPC() {
	//TODO: наверное лучше спрятать этот код в модуль раоты с грпц
	lis, err := net.Listen(config.GRPCNetwork, config.GRPCPort)
	if err != nil {
		log.Fatal(err.Error())
	}
	s := grpc.NewServer()
	serv := admingrpc.AdminServiceServer{}
	service.RegisterAdminServiceServer(s, &serv)
	// Register reflection service on gRPC server.
	serv.Dbs(a.dbs)
	reflection.Register(s)
	a.admin = &serv
	if err := s.Serve(lis); err != nil {
		log.Fatal(err.Error())
	}
}

//ConnectNATS - tries to make connection to NATS
func (a *App) ConnectNATS() {
	//TODO: наверное лучге спрятать этот код в модуль с натсом
	connats := adminnats.NatsService{}
	nc, err := nats.Connect(config.NATSPort)
	if err != nil {
		log.Fatal("connect" + err.Error())
	}
	connats.ConnectionNats(nc)
	a.admin.Nats(&connats)
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
