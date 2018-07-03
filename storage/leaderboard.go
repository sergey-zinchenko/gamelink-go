package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
)

const (
	friendsLeaderboard             = "friends"
	allUsersLeaderboard            = "all"
	friendsAndAllUsersLeaderboards = "full"
)

//LeaderboardString - return leaderboards
func (u User) LeaderboardString(lbType string, lbNum int) (string, error) {
	var result string
	var err error
	var flag int
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	err = u.dbs.mySQL.QueryRow(queries.IternalCheckFlag, u.ID()).Scan(&flag)
	if err != nil {
		return "", err
	}
	if flag == 1 {
		err = graceful.ForbiddenError{Message: "request for deleted user"}
		return "", err
	}
	if lbNum == 1 || lbNum == 2 || lbNum == 3 {
		switch lbType {
		case allUsersLeaderboard:
			err = u.dbs.mySQL.QueryRow(fmt.Sprintf(queries.AllUsersLeaderboardQuery, lbNum), u.ID(), u.ID(), u.ID()).Scan(&result)
		case friendsLeaderboard:
			err = u.dbs.mySQL.QueryRow(fmt.Sprintf(queries.FriendsLeaderboardQuery, lbNum), u.ID(), u.ID(), u.ID()).Scan(&result)
		case friendsAndAllUsersLeaderboards:
			err = u.dbs.mySQL.QueryRow(fmt.Sprintf(queries.FullLeaderboardQuery, lbNum), u.ID(), u.ID(), u.ID(), u.ID(), u.ID()).Scan(&result)
		default:
			return "", graceful.BadRequestError{Message: "wrong leader board type"}
		}
	} else {
		return "", graceful.BadRequestError{Message: "wrong leader board number"}
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return "", graceful.NotFoundError{Message: "user not found"}
		}
		return "", err
	}
	return result, nil
}
