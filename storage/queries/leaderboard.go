package queries

const (
	//AllUsersLeaderboardQuery - mysql query template to get leader board against all users
	AllUsersLeaderboardQuery = `
SELECT JSON_OBJECT(
           "id", 		i.id,
           "nickname", 	i.nickname,
		   "country", 	i.country,
           "score", 	i.score,
		   "rank", 		rank, 
           "meta", 		i.meta,
           "leaderboard", leaderboard) FROM (SELECT * FROM leader_board%[1]d u WHERE u.id=?) as i,
															 (SELECT (CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                                          '"id":', 			l.id, ',',
                                                                          '"nickname":', 	JSON_QUOTE(l.nickname), ',',
                                                                          '"score":', 		l.score, ',',
																		   IFNULL(CONCAT('"country":', JSON_QUOTE(l.country), ','),''),
																		   IFNULL(CONCAT('"meta":', l.meta),''), ',',	
																		  '"country":',		JSON_QUOTE(l.country),
                                                                          '}')), ']'), "[]") AS JSON)) as leaderboard
   FROM
  (SELECT v.id, v.nickname,v.score,v.meta, v.country FROM leader_board%[1]d v WHERE v.score > 0 LIMIT 100) l WHERE  l.id != ?) as q, 
  (SELECT COUNT(*) + 1 as rank FROM leader_board%[1]d i, leader_board%[1]d j WHERE i.id = ? AND j.score > i.score) as rank`

	//FriendsLeaderboardQuery - mysql query template to get leader board against friends
	FriendsLeaderboardQuery = `
SELECT CAST(CONCAT(
    '{"id":', i.id, ','
    '"nickname":', JSON_QUOTE(i.nickname), ','
    '"score":', IFNULL(i.score, 0), ','
    '"rank":', rank, ',',
    IFNULL(CONCAT('"country":', JSON_QUOTE(i.country), ','),''),
    IFNULL(CONCAT('"meta":', i.meta, ','),''),
    '"leaderboard":', leaderboard.leaderboard,'}') AS JSON) as leaderboard
FROM
  (SELECT *
   FROM leader_board%[1]d u
   WHERE u.id = ?) as i,
  (SELECT
     i.*,
     COUNT(*) + 1 as rank
   FROM leader_board%[1]d i, friends f, leader_board%[1]d j
   WHERE i.id =? AND f.user_id1 = i.id AND j.id = f.user_id2 AND j.score > i.score
  ) as rank,
  (SELECT (CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                                '"id":', j.id, ',',
                                                                '"nickname":', JSON_QUOTE(j.nickname), ',',
                                                                '"score":', j.score,',',
                                                                IFNULL(CONCAT('"country":', JSON_QUOTE(j.country), ','),''),
                                                                IFNULL(CONCAT('"meta":', j.meta),''),
                                                                '}')), ']'), "[]") AS JSON)) AS leaderboard
   FROM friends f, leader_board%[1]d j
   WHERE f.user_id1 = ? AND j.id = f.user_id2 AND j.score > 0
  ) as leaderboard`
)
