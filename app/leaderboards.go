package app

import (
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

//getLeaderboard - function to get leaderboard
func (a *App) getLeaderboard(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	lbID, err := ctx.Params().GetInt("id")
	if err != nil {
		handleError(err, ctx)
		return
	}
	leaderboard, err := user.LeaderboardString(ctx.Params().Get("lbtype"), lbID, a.ranks)
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(leaderboard)
}
