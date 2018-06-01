package queries

const (

	//CheckUserQuery - mysql query to check user existence
	CheckUserQuery = `
SELECT id
FROM users u
WHERE u. % s = ?`

	//RegisterUserQuery - mysql query to register user
	RegisterUserQuery = `INSERT INTO users (data) VALUES (?)`

	//GetExtraUserDataQuery - mysql query to get extended user data json
	GetExtraUserDataQuery = `
SET @param = ?;
SELECT IFNULL((SELECT JSON_INSERT(u.data, '$.friends', fj.friends)
               FROM users u,
                 (SELECT CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
                                                                       '\"id\":', b.id,
                                                                       ',', '\"name\":', JSON_QUOTE(b.name),
                                                                       '}')), ']') AS JSON
                         )
                   AS friends
                  FROM
                    (SELECT
                       u.id,
                       u.name,
                       f.user_id2 as g
                     FROM friends f, users u
                     WHERE user_id = @param AND f.user_id1 = u.id
                     UNION
                     SELECT
                       u.id,
                       u.name,
                       f.user_id1 as g
                     FROM friends f, users u
                     WHERE user_id1 = @param AND f.user_id2 = u.id) b
                  GROUP BY b.g) fj
               WHERE u.id = @param), q.data) data
FROM users q
WHERE q.id = @param;`

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
INSERT IGNORE INTO friends (user_id1, user_id2) SELECT
                                                  GREATEST(ids.id1, ids.id2),
                                                  LEAST(ids.id1, ids.id2)
                                                FROM (SELECT
                                                        ?     as id1,
                                                        u2.id as id2
                                                      FROM (SELECT id
                                                            FROM users u
                                                            WHERE u.%s = ?) u2) ids`
)
