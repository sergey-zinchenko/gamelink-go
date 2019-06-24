package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	var err error
	type hasScan interface {
		Scan(dest ...interface{}) error
	}
	//lbUser - struct for lbUser
	type lbUser struct {
		ID       int64  `json:"id"`
		Nickname string `json:"nickname,omitempty"`
		Score    string `json:"score,omitempty"`
		Country  string `json:"country,omitempty"`
		Meta     string `json:"meta,omitempty"`
	}
	//lbResponse - struct for response
	type lbResponse struct {
		lbUser
		Rank        int      `json:"rank"`
		Leaderboard []lbUser `json:"leaderboard"`
	}

	var datacheck = func(lbUser *lbUser, scan hasScan) error {
		var id int64
		var nickname, score, country, meta sql.NullString
		err := scan.Scan(&id, &nickname, &score, &country, &meta)
		if err != nil {
			return err
		}
		lbUser.ID = id
		if nickname.Valid {
			lbUser.Nickname = nickname.String
		}
		if score.Valid {
			lbUser.Score = score.String
		} else {
			lbUser.Score = "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		}
		if country.Valid {
			lbUser.Country = country.String
		}
		if meta.Valid {
			lbUser.Meta = meta.String
		}
		return nil
	}
	row := u.dbs.mySQL.QueryRow(fmt.Sprintf(queries.MyInfoForLeaderboard, lbNum), u.ID())
	res := lbResponse{}
	err = datacheck(&res.lbUser, row)
	if err != nil {
		return "", err
	}
	rows, err := u.dbs.mySQL.Query(fmt.Sprintf(queries.AllUsersLeaderboardQuery, lbNum))
	if err != nil {
		return "", err
	}
	defer rows.Close()
	res.Leaderboard = make([]lbUser, 100)
	var i int
	res.Rank = 0
	for rows.Next() {
		err = datacheck(&res.Leaderboard[i], rows)
		if err != nil {
			return "", err
		}
		if res.Leaderboard[i].ID == u.ID() {
			res.Rank = i + 1
			res.Leaderboard[i] = lbUser{}
			i--
		}
		i++
	}
	res.Leaderboard = res.Leaderboard[0:i]
	if res.Rank == 0 {
		res.Rank = u.dbs.ranks.GetRank(lbNum, res.Score)
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
