package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
)

const (
	friendsLeaderboard  = "friends"
	allUsersLeaderboard = "all"
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
			result, err = u.getAllUsersLeaderboard(lbNum)
		case friendsLeaderboard:
			err = u.dbs.mySQL.QueryRow(fmt.Sprintf(queries.FriendsLeaderboardQuery, lbNum), u.ID(), u.ID(), u.ID()).Scan(&result)
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

//getAllUsersLeaderboard - getting "all" leaderboard
func (u User) getAllUsersLeaderboard(lbNum int) (string, error) {
	var id, rank int
	var nickname, score, country, meta string
	var leaderboard []byte
	var err error
	resMap := make(C.J)
	err = u.dbs.mySQL.QueryRow(fmt.Sprintf(queries.AllUsersLeaderboardQuery, lbNum), u.ID(), u.ID()).Scan(&id, &nickname, &score, &country, &meta, &leaderboard)
	if err != nil {
		return "", err
	}
	resMap["id"] = id
	resMap["nickname"] = nickname
	resMap["score"] = score
	resMap["country"] = country
	resMap["meta"] = meta
	var lb []C.J
	err = json.Unmarshal(leaderboard, &lb)
	resMap["leaderboard"] = lb
	if err != nil {
		return "", err
	}
	rank = u.dbs.ranks.GetRank(lbNum, score)
	resMap["rank"] = rank
	bytes, err := json.Marshal(resMap)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
