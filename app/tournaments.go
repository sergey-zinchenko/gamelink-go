package app

import (
	"gamelink-go/graceful"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"net/http"
	"strconv"
)

//startTournament - func to start tournament from cron
func (a *App) startTournament(ctx iris.Context) {
	var err error
	var usersInRoom, tournamentDuration, registrationDuration int64
	usersInRoom, err = ctx.PostValueInt64("users_in_room")
	tournamentDuration, err = ctx.PostValueInt64("tournament_duration")
	registrationDuration, err = ctx.PostValueInt64("registration_duration")
	if err != nil {
		handleError(err, ctx)
		return
	}
	if usersInRoom < 1 {
		err = graceful.BadRequestError{Message: "wrong count users in room"}
		handleError(err, ctx)
		return
	}
	if tournamentDuration < 1 || registrationDuration < 1 || tournamentDuration < registrationDuration {
		err = graceful.BadRequestError{Message: "wrong tournament or registration duration"}
		handleError(err, ctx)
		return
	}
	err = a.dbs.StartTournament(usersInRoom, tournamentDuration, registrationDuration)
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

//joinTournament - function to join tournament
func (a *App) joinTournament(ctx iris.Context) {
	var err error
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	tournamentID, err := ctx.Params().GetInt("tournament_id")
	if err != nil {
		handleError(err, ctx)
		return
	}
	if tournamentID == 0 {
		err = graceful.BadRequestError{Message: "wrong tournament id"}
		handleError(err, ctx)
		return
	}

	err = user.Join(tournamentID)
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

//updatePts - method to update users pts in tournament
func (a *App) updatePts(ctx iris.Context) {
	var err error
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	tournamentID, err := ctx.Params().GetInt("tournament_id")
	if err != nil {
		handleError(err, ctx)
		return
	}
	s := ctx.Request().URL.Query()["score"][0]
	score, err := strconv.Atoi(s)
	if err != nil {
		handleError(err, ctx)
		return
	}
	err = user.UpdateTournamentScore(tournamentID, score)
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

//getRoomLeaderboard - method to get leaderboard from user tournament room
func (a *App) getRoomLeaderboard(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	tournamentID, err := ctx.Params().GetInt("tournament_id")
	if err != nil {
		handleError(err, ctx)
		return
	}
	leaderboard, err := user.GetLeaderboard(tournamentID)
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(leaderboard)
}

//getAvailiableTournaments - metgod to get available tournaments
func (a *App) getAvailableTournaments(ctx iris.Context) {
	availableTournaments, err := a.dbs.GetTournaments()
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(availableTournaments)
}
