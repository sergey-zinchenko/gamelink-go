package queries

const (
	//AllUsersLeaderboardQuery - mysql query template to get leader board against all users
	AllUsersLeaderboardQuery = `SELECT v.id, v.nickname,v.score,v.meta, v.country FROM leader_board%[1]d v WHERE v.score > 0 LIMIT 100`

	//MyInfoForLeaderboard - mysql query to get user info for leaderboard
	MyInfoForLeaderboard = `SELECT v.id, v.nickname,v.lb%[1]d,v.meta, v.country FROM users v WHERE id = ?`

	//FriendsLeaderboardQuery - mysql query template to get leader board against friends
	FriendsLeaderboardQuery = `
SELECT CAST(CONCAT(
    '{"id":', i.id, ',',
    '"nickname":', JSON_QUOTE(i.nickname), ',',
    '"score":', IFNULL(JSON_QUOTE(i.score), 0), ',',
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
                                                                '"score":', JSON_QUOTE(j.score),
                                                                IFNULL(CONCAT(',"country":', JSON_QUOTE(j.country)),''),
                                                                IFNULL(CONCAT(',"meta":', j.meta),''),
                                                                '}')), ']'), "[]") AS JSON)) AS leaderboard
   FROM friends f, leader_board%[1]d j
   WHERE f.user_id1 = ? AND j.id = f.user_id2 AND j.score > 0
  ) as leaderboard`
)
