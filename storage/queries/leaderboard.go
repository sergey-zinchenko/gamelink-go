package queries

const (
	//AllUsersLeaderboardQuery - mysql query template to get leader board against all users
	AllUsersLeaderboardQuery = `
SELECT JSON_OBJECT(
           "id", 		i.id,
           "nickname", 	IFNULL(i.nickname, i.name),
		   "country", 	i.country,
           "score", 	i.score,
		   "rank", 		rank, 
           "meta", 		i.meta,
           "leaderboard", leaderboard) FROM (SELECT * FROM leader_board%[1]d u WHERE u.id=?) as i,
															 (SELECT (CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                                          '"id":', 			l.id, ',',
                                                                          '"nickname":', 	IFNULL(JSON_QUOTE(l.nickname),JSON_QUOTE(l.name)), ',',
                                                                          '"score":', 		l.score, ',',
																		  '"country":',		JSON_QUOTE(l.country),
                                                                          '}')), ']'), "[]") AS JSON)) as leaderboard
   FROM
  (SELECT v.id, v.nickname,v.name,v.score, v.country FROM leader_board%[1]d v WHERE v.score > 0 LIMIT 100) l WHERE  l.id != ?) as q, 
  (SELECT COUNT(*) + 1 as rank FROM leader_board%[1]d i, leader_board%[1]d j WHERE i.id = ? AND j.score > i.score) as rank`

	//FriendsLeaderboardQuery - mysql query template to get leader board against friends
	FriendsLeaderboardQuery = `
SELECT JSON_OBJECT(
    "id"		,   i.id,
    "nickname"	, 	IFNULL(i.nickname, i.name),
    "score"		,  	IFNULL(i.score, 0),
	"rank"		,   rank,
	"country"	,	i.country,
    "leaderboard", 	leaderboard) FROM (SELECT * FROM leader_board%[1]d u WHERE u.id=?) as i,
				(SELECT (CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                            '"id":',     	l.id,         ',',
                                                            '"nickname":',  IFNULL(JSON_QUOTE(l.nickname),JSON_QUOTE(l.name)), ',',
                                                            '"score":',   	l.score, ',',
															'"country":',	JSON_QUOTE(l.country),
                                                            '}')), ']'),"[]") AS JSON)) AS leaderboard 
FROM 
(SELECT j.id, j.nickname,j.name, j.country, j.score FROM friends f, leader_board%[1]d j WHERE
	((f.user_id1 = ? AND j.id = f.user_id2) OR (f.user_id2 = ? AND j.id = f.user_id1)) AND j.score > 0
) as l) as c,
(SELECT i.*, COUNT(*) + 1 as rank FROM leader_board%[1]d  i, friends f,  leader_board%[1]d  j
WHERE i.id = ?
AND 
((f.user_id1 = i.id AND j.id = f.user_id2) OR (f.user_id2 = i.id AND j.id = f.user_id1)) 
AND
j.score > i.score) as rank`
)
