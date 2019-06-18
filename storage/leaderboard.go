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
func (u User) LeaderboardString(lbType string, lbNum int, ranks *Ranks) (string, error) {
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
			result, err = u.getAllUsersLeaderboard(lbNum, ranks)
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
func (u User) getAllUsersLeaderboard(lbNum int, ranks *Ranks) (string, error) {
	if ranks == nil {
		return "", graceful.ServiceUnavailableError{Message: "ranks pointer is nil"}
	}
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
	switch lbNum {
	case 1:
		rank, err = getRank(score, ranks.RankArr1)
	case 2:
		rank, err = getRank(score, ranks.RankArr1)
	case 3:
		rank, err = getRank(score, ranks.RankArr1)
	}
	if err != nil {
		return "", err
	}
	resMap["rank"] = rank
	resultString, err := json.Marshal(resMap)
	if err != nil {
		return "", err
	}
	return string(resultString), nil
}

//getRank - getting user rank using binsearch
func getRank(score string, rankArr *[]string) (int, error) {
	var rank int
	if rankArr == nil {
		return 0, graceful.ServiceUnavailableError{Message: "rankArr pointer is nil"}
	}
	arr := *rankArr
	start := 0
	end := len(arr) - 1
	if score >= arr[start] {
		return 1, nil
	}
	if score <= arr[end] {
		return end + 1, nil
	}
	for {
		median := (start + end) / 2
		if score > arr[median] {
			if score < arr[median-1] {
				rank = median
				break
			}
			end = median
		} else if score < arr[median] {
			if score > arr[median+1] {
				rank = median + 1
				break
			}
			start = median
		} else {
			rank = median + 1
			break
		}
	}
	return rank, nil
}
