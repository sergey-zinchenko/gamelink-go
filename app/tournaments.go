package app

import (
	C "gamelink-go/common"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"net/http"
)

//startTournament - func to start tournament from cron
func (a *App) startTournament(ctx iris.Context) {
	var err error
	err = a.dbs.StartTournament()
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.StatusCode(http.StatusOK)
}

//joinTournament - function to join tournament
func (a *App) joinTournament(ctx iris.Context) {
	var err error
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	err = user.Join()
	if err != nil {
		handleError(err, ctx)
		return
	}
}

//updatePts - method to update users pts in tournament
func (a *App) updatePts(ctx iris.Context) {
	var (
		data C.J
		err  error
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	err = ctx.ReadJSON(&data)
	if err != nil {
		return
	}
	err = user.UpdateTournamentScore(data)
	if err != nil {
		return
	}
}

//getRoomLeaderboard - method to get leaderboard from user tournament room
func (a *App) getRoomLeaderboard(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	leaderboard, err := user.GetLeaderboard()
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(leaderboard)
}
