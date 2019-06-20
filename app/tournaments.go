package app

import (
	"gamelink-go/graceful"
	"gamelink-go/storage"
	"github.com/gorhill/cronexpr"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	lbRegexp *regexp.Regexp
	err      error
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	lbRegexp, err = regexp.Compile("^\\d{100}$")
	if err != nil {
		log.Fatal(err)
	}
}

//startTournament - func to start tournament from cron
func (a *App) startTournament(ctx iris.Context) {
	var err error
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	var getUsersInRoom, getTournamentDuration, getRegistrationDuration []string

	getUsersInRoom = ctx.Request().URL.Query()["users_in_room"]
	if getUsersInRoom == nil || getUsersInRoom[0] == "" {
		err = graceful.BadRequestError{Message: "invalid param users in room"}
		return
	}

	getTournamentDuration = ctx.Request().URL.Query()["tournament_duration"]
	if getTournamentDuration == nil || getTournamentDuration[0] == "" {
		err = graceful.BadRequestError{Message: "invalid tournament duration"}
		return
	}

	getRegistrationDuration = ctx.Request().URL.Query()["registration_duration"]
	if getRegistrationDuration == nil || getRegistrationDuration[0] == "" {
		err = graceful.BadRequestError{Message: "invalid registration duration"}
		return
	}
	usersInRoom, err := strconv.ParseInt(getUsersInRoom[0], 10, 64)
	if err != nil {
		return
	}
	if usersInRoom < 1 {
		err = graceful.BadRequestError{Message: "wrong users count in room"}
		return
	}
	tournamentDuration, err := strconv.ParseInt(getTournamentDuration[0], 10, 64)
	if err != nil {
		return
	}
	registrationDuration, err := strconv.ParseInt(getRegistrationDuration[0], 10, 64)
	if err != nil {
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
	err = tournament.Join(user.ID())
	if err != nil {
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
	score := ctx.PostValue("score")
	matched := lbRegexp.MatchString(score)
	if err != nil {
		return
	}
	if !matched {
		err = graceful.BadRequestError{Message: "wrong score"}
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

func (a App) timeToNextTournament(ctx iris.Context) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	out := strings.Split(string(output), "\n")
	var tournamentTask string
	for _, v := range out {
		if strings.Contains(v, "tournament.sh") {
			tournamentTask = v
			break
		}
	}
	if tournamentTask == "" {
		return
	}
	r, err := regexp.Compile("^(((([*])|(((([0-5])?[0-9])((-(([0-5])?[0-9])))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))(,(((([*])|(((([0-5])?[0-9])((-(([0-5])?[0-9])))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?)))* (((([*])|(((((([0-1])?[0-9]))|(([2][0-3])))((-(((([0-1])?[0-9]))|(([2][0-3])))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))(,(((([*])|(((((([0-1])?[0-9]))|(([2][0-3])))((-(((([0-1])?[0-9]))|(([2][0-3])))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?)))* (((((((([*])|(((((([1-2])?[0-9]))|(([3][0-1]))|(([1-9])))((-(((([1-2])?[0-9]))|(([3][0-1]))|(([1-9])))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))|(L)|(((((([1-2])?[0-9]))|(([3][0-1]))|(([1-9])))W))))(,(((((([*])|(((((([1-2])?[0-9]))|(([3][0-1]))|(([1-9])))((-(((([1-2])?[0-9]))|(([3][0-1]))|(([1-9])))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))|(L)|(((((([1-2])?[0-9]))|(([3][0-1]))|(([1-9])))W)))))*)|([?])) (((([*])|((((([1-9]))|(([1][0-2])))((-((([1-9]))|(([1][0-2])))))?))|((((JAN)|(FEB)|(MAR)|(APR)|(MAY)|(JUN)|(JUL)|(AUG)|(SEP)|(OKT)|(NOV)|(DEC))((-((JAN)|(FEB)|(MAR)|(APR)|(MAY)|(JUN)|(JUL)|(AUG)|(SEP)|(OKT)|(NOV)|(DEC))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))(,(((([*])|((((([1-9]))|(([1][0-2])))((-((([1-9]))|(([1][0-2])))))?))|((((JAN)|(FEB)|(MAR)|(APR)|(MAY)|(JUN)|(JUL)|(AUG)|(SEP)|(OKT)|(NOV)|(DEC))((-((JAN)|(FEB)|(MAR)|(APR)|(MAY)|(JUN)|(JUL)|(AUG)|(SEP)|(OKT)|(NOV)|(DEC))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?)))* (((((((([*])|((([0-6])((-([0-6])))?))|((((SUN)|(MON)|(TUE)|(WED)|(THU)|(FRI)|(SAT))((-((SUN)|(MON)|(TUE)|(WED)|(THU)|(FRI)|(SAT))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))|((([0-6])L))|(W)|(([#][1-5]))))(,(((((([*])|((([0-6])((-([0-6])))?))|((((SUN)|(MON)|(TUE)|(WED)|(THU)|(FRI)|(SAT))((-((SUN)|(MON)|(TUE)|(WED)|(THU)|(FRI)|(SAT))))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))|((([0-6])L))|(W)|(([#][1-5])))))*)|([?]))((( (((([*])|((([1-2][0-9][0-9][0-9])((-([1-2][0-9][0-9][0-9])))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?))(,(((([*])|((([1-2][0-9][0-9][0-9])((-([1-2][0-9][0-9][0-9])))?)))((/(([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?([0-9])?[0-9])))?)))*))?)")
	if err != nil {
		return
	}
	matched := r.FindStringSubmatch(tournamentTask)
	nextTime := cronexpr.MustParse(matched[0]).Next(time.Now()).Unix() - time.Now().Unix()
	ctx.WriteString(strconv.FormatInt(nextTime, 10))
}
