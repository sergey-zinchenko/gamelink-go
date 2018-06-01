package storage

import (
	"database/sql"
	"errors"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
)

const (
	friendsLeaderboard  = "friends"
	allUsersLeaderboard = "all"
)

//Leaderboard - return leaderboards
func (u User) Leaderboard(lbType string, lbNum int) (string, error) {
	var result string
	var err error
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	switch lbType {
	case allUsersLeaderboard:
		switch lbNum {
		case 1:
			err = u.dbs.mySQL.QueryRow(queries.AllUsersLeaderboard1Query, u.ID()).Scan(&result)
		case 2:
			err = u.dbs.mySQL.QueryRow(queries.AllUsersLeaderboard2Query, u.ID()).Scan(&result)
		default:
			return "", graceful.BadRequestError{Message: "wrong leader board number"}
		}
	case friendsLeaderboard:
		switch lbNum {
		case 1:
			err = u.dbs.mySQL.QueryRow(queries.FriendsLeaderboard1Query, u.ID(), u.ID(), u.ID(), u.ID(), u.ID()).Scan(&result)
		case 2:
			err = u.dbs.mySQL.QueryRow(queries.FriendsLeaderboard2Query, u.ID(), u.ID(), u.ID(), u.ID(), u.ID()).Scan(&result)
		default:
			return "", graceful.BadRequestError{Message: "wrong leader board number"}
		}
	default:
		return "", graceful.BadRequestError{Message: "wrong leader board type"}
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return "", graceful.NotFoundError{Message: "user not found"}
		}
		return "", err
	}
	return result, nil
}
