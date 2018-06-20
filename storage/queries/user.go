package queries

const (

	//CheckUserQuery - mysql query to check user existence
	CheckUserQuery = `
SELECT id
FROM users u
WHERE u. %s = ?`

	//RegisterUserQuery - mysql query to register user
	RegisterUserQuery = `INSERT INTO users (data) VALUES (?)`

	//GetExtraUserDataQuery - mysql query to get extended user data json
	GetExtraUserDataQuery = `
SELECT IFNULL((SELECT JSON_INSERT(u.data, 
								'$.friends', fj.friends, 
                                '$.saves', 	w.saves,
                                '$.tournaments', q.tournaments)
               FROM users u, (SELECT CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{','"id":', s.id,'}')), ']') AS JSON) as saves FROM (SELECT id from saves WHERE user_id = ?)s) w,
               (SELECT CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{','"tournament_id":', t.tournament_id,'}')), ']') AS JSON) as tournaments FROM (SELECT tournament_id from users_tournaments WHERE user_id = ?)t) q,
                 (SELECT CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                                       '"id":', b.id,
                                                                       ',', '"name":', JSON_QUOTE(b.name),
                                                                       '}')), ']') AS JSON) AS friends
                  FROM
                    (SELECT
                       u.id,
                       u.name,
                       f.user_id2 as g
                     FROM friends f, users u
                     WHERE user_id2 = ? AND f.user_id1 = u.id
                     UNION
                     SELECT
                       u.id,
                       u.name,
                       f.user_id1 as g
                     FROM friends f, users u
                     WHERE user_id1 = ? AND f.user_id2 = u.id) b
                  GROUP BY b.g) fj
               WHERE u.id = ?), q.data) data
FROM users q
WHERE q.id = ?;`

	//GetUserDataQuery - mysql query to get user's data json
	GetUserDataQuery = `
SELECT data
FROM users u
WHERE u.id = ?`

	//UpdateUserDataQuery - mysql query to update data field of the user record
	UpdateUserDataQuery = `
UPDATE users u
SET u.data = ?
WHERE u.id = ?`

	//DeleteUserQuery - mysql query to delete user
	DeleteUserQuery = `
DELETE FROM users
WHERE id = ?`

	//MakeFriendshipQuery - mysql query to make friendship between two users
	MakeFriendshipQuery = `
INSERT IGNORE INTO friends (user_id1, user_id2) VALUES 
(?, (SELECT id from users u WHERE u.%[1]s =?)),
((SELECT id from users u WHERE u.%[1]s =?), ?)`
)
