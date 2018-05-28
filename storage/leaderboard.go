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
		q := fmt.Sprintf("SELECT JSON_OBJECT("+
			"\"id\"       ,  k.`id`   , "+
			"\"position\" ,  k.`pos`  , "+
			"\"name\"     ,  k.`name` , "+
			"\"score\"    ,  k.`score`, "+
			"\"top100\"   , "+
			"CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{',"+
			"'\"id\":'    , 	b.`id`, 				','  ,"+
			"'\"name\":'  , 	JSON_QUOTE(b.`name`),   ','  ,"+
			"'\"score\":' ,  	b.`%s`						 ,"+
			"'}')),']') AS JSON))"+
			"FROM(SELECT s.`id`, count(*) + 1 as pos, s.`name`, s.`score` from leader_board1 l, (select id, %s as score, name from leader_board1 o where o.`id` = %d) s "+
			"where l.`%s` IS NOT NULL AND  l.`%s` > s.`score` ) k,"+
			"(SELECT l.`id`, l.`name`, l.`%s` FROM leader_board1 l LIMIT 100) b GROUP BY k.`id`", lbName, lbName, u.ID(), lbName, lbName, lbName)
		err = u.dbs.mySQL.QueryRow(q).Scan(&lb)
	} else if lbType == "friends" {
		q := fmt.Sprintf("SELECT JSON_OBJECT("+
			"\"id\"			,  k.`id`		, "+
			"\"position\"	,  k.`pos`		, "+
			"\"name\"		,  k.`name`		, "+
			"\"score\"		,  k.`score`	, "+
			"\"friends\"	,"+
			"CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{',"+
			"'\"id\":'	  , 	p.`id`, 				',' ,"+
			"'\"name\":'  , 	JSON_QUOTE(p.`name`),   ',' ,"+
			"'\"score\":' , 	p.`%s`,"+
			"'}')),']') AS JSON))"+
			"FROM ( SELECT v.`id`, v.`name`, v.`%s`"+
			"FROM (SELECT u.`id`,u.`name`,u.`%s` FROM `friends` f, `users` u WHERE f.`user_id2` = %d AND f.`user_id1` = u.`id` "+
			"UNION SELECT u.`id`,u.`name`,u.`%s` FROM `friends` f, `users` u WHERE f.`user_id1` = %d AND f.`user_id2` = u.`id`) v "+
			"ORDER BY v.`%s`) p, "+
			"(SELECT m.`id`, count(*) + 1 as pos, m.`name`, m.score "+
			"FROM (SELECT  l.`id`, l.`name`, l.`%s` as score FROM leader_board1 l WHERE l.`id` = %d) m, "+
			"(SELECT u.`%s` as score FROM `friends` f, `users` u WHERE f.`user_id2` = %d AND f.`user_id1` = u.`id` "+
			"UNION "+
			"SELECT u.`%s` as score FROM `friends` f, `users` u WHERE f.`user_id1` = %d AND f.`user_id2` = u.`id`) s "+
			"where m.score IS NOT NULL AND  s.score > m.score ) k GROUP BY k.`id`", lbName, lbName, lbName, u.ID(), lbName, u.ID(), lbName, lbName, u.ID(), lbName, u.ID(), lbName, u.ID())
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
