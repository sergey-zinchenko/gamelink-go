package storage

import (
	"database/sql"
	"fmt"
	"gamelink-go/storage/queries"
	"github.com/kataras/iris/core/errors"
	"sort"
	"sync"
)

//Ranks - struct to work with rank arrays. Count of rank arrays == count of leaderboards
type Ranks struct {
	ranks []*Rank
}

//Rank - struct to work with leaderboard rank
type Rank struct {
	mysql   *sql.DB
	num     int
	rankArr *[]string
	mutex   sync.RWMutex
}

//MakeRank - Rank object constructor
func MakeRank(mysql *sql.DB, num int) *Rank {
	return &Rank{mysql: mysql, num: num}
}

//Fill - fill rankArr from db
func (r *Rank) Fill(lbNum int) error {
	if lbNum <= 0 || lbNum > NumOfLeaderBoards {
		return errors.New("wrong leaderboard number")
	}
	var res []string
	q := fmt.Sprintf(queries.GetUserScoresFromDb, lbNum)
	rows, err := r.mysql.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r string
		rows.Scan(&r)
		res = append(res, r)
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.rankArr = &res
	return nil
}

//GetRank - return user rank in leaderboard
func (r *Rank) GetRank(score string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	l := len(*r.rankArr)
	indexInArray := sort.Search(l, func(i int) bool { return (*r.rankArr)[i] >= score })
	return indexInArray + 1
}

//MakeRanks - constructor for Ranks
func MakeRanks(num int, mysql *sql.DB) *Ranks {
	var result = &Ranks{ranks: make([]*Rank, num)}
	for i := 0; i < num; i++ {
		result.ranks[i] = MakeRank(mysql, i+1)
	}
	return result
}

//Update - update ranks from db
func (r *Ranks) Update() error {
	for k, v := range r.ranks {
		err := v.Fill(k + 1)
		if err != nil {
			return err
		}
	}
	return nil
}

//GetRank - get user rank
func (r *Ranks) GetRank(lbNum int, score string) int {
	return r.ranks[lbNum-1].GetRank(score)
}
