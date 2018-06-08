package app

import (
	"gamelink-go/storage"
	"github.com/kataras/iris"
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

	//ctx.ContentType(context.ContentJSONHeaderValue)
	//ctx.WriteString(leaderboard)
}
