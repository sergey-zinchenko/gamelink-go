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
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	var usersInRoom, tournamentDuration, registrationDuration int64
	usersInRoom, err = ctx.PostValueInt64("users_in_room")
	if err != nil {
		return
	}
	tournamentDuration, err = ctx.PostValueInt64("tournament_duration")
	if err != nil {
		return
	}
	registrationDuration, err = ctx.PostValueInt64("registration_duration")
	if err != nil {
		return
	}
	if usersInRoom < 1 {
		err = graceful.BadRequestError{Message: "wrong users count in room"}
		return
	}
	if tournamentDuration < 60 || registrationDuration < 60 || tournamentDuration < registrationDuration {
		err = graceful.BadRequestError{Message: "wrong tournament or registration duration"}
		return
	}
	err = a.dbs.StartTournament(usersInRoom, tournamentDuration, registrationDuration)
	if err != nil {
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
	tournament, err := a.dbs.Tournament(tournamentID)
	if err != nil {
		handleError(err, ctx)
		return
	}
	err = tournament.Join(user.ID())
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

//updatePts - method to update users pts in tournament
func (a *App) updateScore(ctx iris.Context) {
	var err error
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	tournamentID, err := ctx.Params().GetInt("tournament_id")
	if err != nil {
		return
	}
	tournament, err := a.dbs.Tournament(tournamentID)
	if err != nil {
		return
	}
	s := ctx.Request().URL.Query()["score"][0]
	score, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	err = tournament.UpdateTournamentScore(user.ID(), score)
	if err != nil {
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
	tournament, err := a.dbs.Tournament(tournamentID)
	if err != nil {
		handleError(err, ctx)
		return
	}
	leaderboard, err := tournament.GetLeaderboard(user.ID())
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(leaderboard)
}

//getAvailiableTournaments - metgod to get available tournaments
func (a *App) getAvailableTournaments(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	availableTournaments, err := user.GetTournaments()
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(availableTournaments)
}

//getUsersResults - method to get all user results in last 100 tournaments
func (a *App) getUsersResults(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	availableTournaments, err := user.GetResults()
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(availableTournaments)
}
