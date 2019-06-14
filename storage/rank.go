package storage

import (
	"errors"
	"fmt"
)

//Ranks - struct to work with rank arrays. Count of rank arrays == count of leaderboards
type Ranks struct {
	DBS      *DBS
	RankArr1 *[]string
	RankArr2 *[]string
	RankArr3 *[]string
}

//GenerateRankArrays - generate arrays using generate func
func (r *Ranks) GenerateRankArrays(num int) error {
	if num == 0 {
		for i := 1; i <= NumOfLeaderBoards; i++ {
			err := r.generate(i)
			if err != nil {
				return err
			}
		}
	} else if num > 0 && num <= NumOfLeaderBoards {
		err := r.generate(num)
		if err != nil {
			return err
		}
	} else {
		return errors.New("wrong rank array num")
	}
	return nil
}

//generate - get data from db and put it in rank arrays
func (r *Ranks) generate(num int) error {
	var res []string
	q := fmt.Sprintf("SELECT score from gamelink.leader_board%d", num)
	rows, err := r.DBS.mySQL.Query(q)
	if err != nil {
		return nil
	}
	for rows.Next() { //проверить, если errNoRows, вернется ли пустой массив или вылезет ошибка?
		var r string
		rows.Scan(&r)
		res = append(res, r)
	}
	switch n := num; n {
	case 1:
		r.RankArr1 = &res
	case 2:
		r.RankArr2 = &res
	case 3:
		r.RankArr3 = &res
	}
	return nil
}
