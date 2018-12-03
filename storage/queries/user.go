package queries

const (

	//CheckUserQuery - mysql query to check user existence
	CheckUserQuery = `SELECT id FROM users u WHERE u. %s = ?`

	//UpdateSexCountryBdate - query to update fields if it's not in db
	UpdateSexCountryBdate = `UPDATE users u SET u.data = (SELECT JSON_MERGE_PATCH(?, (SELECT data FROM(SELECT k.data from users k WHERE k.id = ?)k))) WHERE u.id = ?`

	//CheckFlag - mysql query to check if user deleted when login
	CheckFlag = `SELECT deleted FROM users u WHERE u.%s = ?`

	//IternalCheckFlag - mysql query to iternal check if user deleted
	IternalCheckFlag = `SELECT deleted FROM users u WHERE u.id = ?`

	//RegisterUserQuery - mysql query to register user
	RegisterUserQuery = `INSERT INTO users (data) VALUES (?)`

	//GetExtraUserDataQuery - mysql query to get extended user data json
	GetExtraUserDataQuery = `
SELECT (SELECT  JSON_MERGE_PATCH (CAST(CONCAT( 	'{"id":'  , 	u.id,
								IFNULL(CONCAT(',"rank1":'  , 	rank1),""),
                                IFNULL(CONCAT(',"rank2":'  , 	rank2),""),
                                IFNULL(CONCAT(',"rank3":'  , 	rank3),""),',',
					 			'"friends":',   	IFNULL(q.friends,"[]"), ',',
                                '"saves":'  , 		IFNULL(w.saves, "[]"), ',',
                                '"tournaments":', IFNULL(v.tournaments, "[]"),
                                '}') AS JSON), IFNULL((select data from users where id=?), "{}"))  
               FROM (SELECT id FROM users  WHERE id = ?) AS u,
               (SELECT CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
														'"id":',   b.id, ',' ,
                                                        '"nickname":', 	JSON_QUOTE(IFNULL(b.nickname,b.name)),
																		'}')), ']') AS JSON) as friends
			   FROM (SELECT u.id, u.nickname, u.name FROM friends f, users u WHERE user_id2 = ? AND f.user_id1 = u.id
				UNION
				SELECT u.id, u.nickname, u.name FROM friends f, users u WHERE user_id1 = ? AND f.user_id2 = u.id) b) q,
                (SELECT CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{',
							'"id":', 	s.id, ',', 
                            '"name":', 	IFNULL(JSON_QUOTE(s.name), ""), ',', 
                            '"state":', s.state, ',', 
                            '"date":',	s.updated_at,
                            '}')), ']') AS JSON) as saves FROM (SELECT id, name, state, UNIX_TIMESTAMP(updated_at) as updated_at from saves WHERE user_id = ?)s) w,
                (SELECT CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{','"id":', t.tournament_id,'}')), ']') AS JSON) as tournaments FROM (SELECT tournament_id from users_tournaments WHERE user_id = ?)t) v,
				(SELECT COUNT(*) + 1 as rank1 FROM leader_board1 i, leader_board1 j WHERE i.id = ? AND j.score > i.score) as rank1,
                (SELECT COUNT(*) + 1 as rank2 FROM leader_board2 i, leader_board2 j WHERE i.id = ? AND j.score > i.score) as rank2,
                (SELECT COUNT(*) + 1 as rank3 FROM leader_board3 i, leader_board3 j WHERE i.id = ? AND j.score > i.score) as rank3) AS userdata`

	//GetUserDataQuery - mysql query to get user's data json
	GetUserDataQuery = `
SELECT data from users u where u.id=? AND u.deleted != 1`

	//UpdateUserDataQuery - mysql query to update data field of the user record
	UpdateUserDataQuery = `
UPDATE users u
SET u.data = ?
WHERE u.id = ? AND u.deleted != 1`

	//UpdateRecoveryUserDataAndIDQuery - mysql query to update data on addAuth when third party user already in database
	UpdateRecoveryUserDataAndIDQuery = `UPDATE users u SET u.id = ?, u.data = ? WHERE u.%s = ? AND u.deleted != 1`

	//DeleteDummyUserFromDB - delete user from table users. USED to delete dummy users only!!!
	DeleteDummyUserFromDB = `DELETE FROM users WHERE id = ? AND dummy = 1;`

	//DeleteUserQuery - mysql query to mark deleted user
	DeleteUserQuery = `
	UPDATE users u SET u.deleted=1 WHERE u.id=?`

	//DeleteUserFromFriends - mysql query to delete user from table friends
	DeleteUserFromFriends = `DELETE FROM friends WHERE user_id1 = ? OR user_id2 = ?`

	//MakeFriendshipQuery - mysql query to make friendship between two users
	MakeFriendshipQuery = `
INSERT IGNORE INTO friends (user_id1, user_id2) VALUES 
(?, (SELECT id from users u WHERE u.%[1]s =? AND u.deleted != 1)),
((SELECT id from users u WHERE u.%[1]s =? AND u.deleted != 1), ?)`

	//AddDeviceID - mysql query to add device id in first user auth
	AddDeviceID = `INSERT IGNORE INTO device_ids (user_id, device_id, message_system) values (?,?,?)`

	//GetUserName - mysql query to get user name
	GetUserName = `SELECT IFNULL(nickname, name) FROM users WHERE id = ?`
)
