package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

//Leaderboard - return leaderboards
func (u User) Leaderboard(lbType string, lbName string) (string, error) {
	var lb string
	var err error
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	if lbType == "all" {
		q := fmt.Sprintf("SET @row_num = 0; SELECT CAST(CONCAT('[', GROUP_CONCAT(CONCAT('{', '\"pos\":',  k.`num`, ',' ,'\"name\":', JSON_QUOTE(k.`name`), ',', '\"score\":',	k.`%s`, '}')),']') AS JSON) "+
			"FROM (SELECT p.`num`, p.`name`, p.`%s` "+
			"FROM (SELECT (@row_num := @row_num +1) As num, u.`name`,u.`%s`, u.`id` FROM `users` u  ORDER BY u.`%s` DESC) p "+
			"WHERE p.`num`<= 100 OR p.`id` = %d) k ", lbName, lbName, lbName, lbName, u.ID())
		fmt.Println(q)
		err = u.dbs.mySQL.QueryRow(q).Scan(&lb)
	} else if lbType == "friends" {
		q := fmt.Sprintf("SELECT CAST(CONCAT('[', GROUP_CONCAT(CONCAT('{', '\"name\":', JSON_QUOTE(p.`name`), ',', '\"score\":',	p.`%s`, '}')),']') AS JSON) "+
			"FROM ( SELECT v.`name`, v.`%s` "+
			"FROM (SELECT u.`name`,u.`%s` FROM `friends` f, `users` u WHERE f.`user_id2` = %d AND f.`user_id1` = u.`id` "+
			"UNION "+
			"SELECT u.`name`,u.`%s` FROM `friends` f, `users` u WHERE f.`user_id1` = %d AND f.`user_id2` = u.`id` "+
			"UNION "+
			"SELECT u.`name`, u.`%s` FROM `users` u WHERE u.`id` = %d) v "+
			"ORDER BY v.`lb1` DESC LIMIT 100) p", lbName, lbName, lbName, u.ID(), lbName, u.ID(), lbName, u.ID())
		err = u.dbs.mySQL.QueryRow(q).Scan(&lb)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("score not found")
		}
		return "", err
	}
	return lb, nil
}
