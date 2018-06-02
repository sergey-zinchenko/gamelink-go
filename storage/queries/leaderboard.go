package queries

const (
	//AllUsersLeaderboardQuery - mysql query template to get leader board against all users
	AllUsersLeaderboardQuery = `
SELECT JSON_OBJECT(
           "id", k.id,
           "name", k.name,
           "score", k.score,
           "rank", k.rank,
           "leaderboard", CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                                          '"id":', l.id, ',',
                                                                          '"name":', JSON_QUOTE(l.name), ',',
                                                                          '"score":', l.score
           ,
                                                                          '}')), ']'), "[]") AS JSON)) AS leaderboard 
   FROM
  (SELECT
     i.*,
     COUNT(*) + 1 as rank
   FROM leader_board%[1]d i, leader_board%[1]d j
   WHERE i.id = ? AND j.score > i.score) k, (SELECT *
                                              FROM leader_board%[1]d WHERE leader_board%[1]d.score > 0
                                              LIMIT 100) l
WHERE
k.id != l.id`

	//FriendsLeaderboardQuery - mysql query template to get leader board against friends
	FriendsLeaderboardQuery = `
SELECT JSON_OBJECT(
    "id",     k.id,
    "name",   k.name,
    "score",  k.score,
    "rank",   k.rank,
    "leaderboard", CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                            '"id":',     l.id,         ',',
                                                            '"name":',   JSON_QUOTE(l.name), ',',
                                                            '"score":',   l.score
                                                            ,
                                                            '}')), ']'),"[]") AS JSON)) AS leaderboard FROM 
(SELECT i.*, COUNT(*) + 1 as rank FROM leader_board%[1]d  i, friends f,  leader_board%[1]d  j
WHERE i.id = ?
AND 
((f.user_id1 = i.id AND j.id = f.user_id2) OR (f.user_id2 = i.id AND j.id = f.user_id1)) 
AND
j.score > i.score) k, friends f1, leader_board%[1]d  l where ((f1.user_id1 = k.id AND l.id = f1.user_id2) OR (f1.user_id2 = k.id AND l.id = f1.user_id1)) AND l.score > 0`
)
