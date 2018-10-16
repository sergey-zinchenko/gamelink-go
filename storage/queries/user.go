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
SELECT (SELECT CAST(CONCAT( 	'{"id":'  , 	u.id, 
								IFNULL(CONCAT(',"vk_id":'    , 	JSON_QUOTE(u.vk_id)),""), 
                                IFNULL(CONCAT(',"fb_id":'    , 	JSON_QUOTE(u.fb_id)),""),
                                IFNULL(CONCAT(',"name":'  	 , 	JSON_QUOTE(u.name)),""),
                                IFNULL(CONCAT(',"nickname":' ,  JSON_QUOTE(u.nickname)),""),
								IFNULL(CONCAT(',"sex":'  	 , 	JSON_QUOTE(u.sex)),""),
								IFNULL(CONCAT(',"email":'    , 	JSON_QUOTE(u.email)),""),
                                IFNULL(CONCAT(',"lb1":'  , 	u.lb1),""), 
                                IFNULL(CONCAT(',"lb2":'  , 	u.lb2),""),
                                IFNULL(CONCAT(',"lb3":'  , 	u.lb3),""),
								IFNULL(CONCAT(',"rank1":'  , 	rank1),""),
                                IFNULL(CONCAT(',"rank2":'  , 	rank2),""),
                                IFNULL(CONCAT(',"rank3":'  , 	rank3),""),
                                IFNULL(CONCAT(',"bdate":'  , 	u.bdate),""), 
                                IFNULL(CONCAT(',"meta":'   , 	u.meta),""),
                                IFNULL(CONCAT(',"country":'  , 	JSON_QUOTE(u.country)),""),',',
					 			'"friends":',   	IFNULL(q.friends,"[]"), ',',
                                '"saves":'  , 		IFNULL(w.saves, "[]"), ',',
                                '"tournaments":', IFNULL(v.tournaments, "[]"),
                                '}') AS JSON)) AS userdata
               FROM (SELECT id, vk_id, fb_id, name, nickname, sex, lb1, lb2, lb3, email, bdate, meta, country FROM users  WHERE id = ?) AS u,
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
                (SELECT COUNT(*) + 1 as rank3 FROM leader_board3 i, leader_board3 j WHERE i.id = ? AND j.score > i.score) as rank3`

	//GetUserDataQuery - mysql query to get user's data json
	GetUserDataQuery = `
SELECT IF((vk_id IS NOT NULL or fb_id IS NOT NULL), 0, 1) as dummy , data from users u where u.id=? AND u.deleted != 1`

	//UpdateUserDataQuery - mysql query to update data field of the user record
	UpdateUserDataQuery = `
UPDATE users u
SET u.data = ?
WHERE u.id = ? AND u.deleted != 1`

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
	AddDeviceID = `INSERT IGNORE INTO device_ids (user_id, device_id, device_os) values (?,?,?)`
)
