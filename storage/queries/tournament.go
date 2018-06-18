package queries

const (
	//SelectMaxExpiredTime - query to select max expired time
	SelectMaxExpiredTime = `SELECT IFNULL((SELECT MAX(expired_time) FROM tournaments),0)`

	//CreateNewTournament - query to create new tournament
	CreateNewTournament = `INSERT INTO tournaments (expired_time) VALUES (?)`

	//CreateNewRoom - query to create new tournament room
	CreateNewRoom = `INSERT INTO rooms (expired_time) VALUES (?)`

	//JoinTournament - query to add user id in table users_tournaments to allow us to check if user already in tournament
	JoinTournament = `INSERT INTO users_tournaments (tournament_id, user_id) VALUES ((SELECT MAX(id) FROM tournaments), ?)`

	//GetCountUsersInRoomAndTournamentExpiredTime - query to get count of users in room to allow us to check max count users in room in current tournament
	GetCountUsersInRoomAndTournamentExpiredTime = `SELECT t.expired_time, c.users_count  FROM 
		(SELECT MAX(expired_time) as expired_time FROM tournaments) as t,
		(SELECT IFNULL(count(user_id),0) as users_count FROM rooms_users WHERE room_id = (SELECT MAX(room_id) FROM rooms_users)) as c`

	//JoinUserToExistRoom - query to join user in existed room
	JoinUserToExistRoom = `INSERT INTO rooms_users (room_id,expired_time, user_id) VALUES ((SELECT MAX(id) FROM rooms),(SELECT MAX(expired_time) FROM tournaments), ?)`

	//CreateNewRoomInCurrentTournament - query to create new room if there max users in last created room
	CreateNewRoomInCurrentTournament = `INSERT INTO rooms (expired_time) VALUES ((SELECT MAX(expired_time) FROM tournaments))`

	//JoinNewRoom - query to join user in created room
	JoinNewRoom = `INSERT INTO rooms_users (room_id,expired_time, user_id) VALUES ((SELECT MAX(id) FROM rooms),(SELECT MAX(expired_time) FROM tournaments), ?)`

	//UpdateUserTournamentScore - query to update user tournament score
	UpdateUserTournamentScore = `UPDATE rooms_users SET score = ? WHERE (SELECT MAX(expired_time)) > ? AND user_id = ?`

	//GetRoomLeaderboard - query to get leaderboard from current user room in current tournament
	GetRoomLeaderboard = `SELECT CAST(CONCAT(
    '{"id":', i.id, ',',
    '"nickname":', JSON_QUOTE(IFNULL(i.nickname,i.name)), ',',
    '"score":', IFNULL(score, 0), ',',
    '"rank":', rank, ',',
    IFNULL(CONCAT('"country":', JSON_QUOTE(i.country), ','),''),
    IFNULL(CONCAT('"meta":', i.lbmeta, ','),''),
    '"leaderboard":', leaderboard,'}') AS JSON) as leaderboard 
           FROM (SELECT u.id, u.name, u.nickname, u.country,u.lbmeta FROM users u WHERE u.id=?) as i,
					 (SELECT (CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
																		  '"id":', 			l.id, ',',
                                                                          '"nickname":', 	JSON_QUOTE(IFNULL(l.nickname,l.name)), ',',
                                                                          '"score":', 		IFNULL(l.score, 0),
																		   IFNULL(CONCAT(',"country":', JSON_QUOTE(l.country)),''),
																		   IFNULL(CONCAT(',"meta":', l.lbmeta),''),
                                                                          '}')), ']'), "[]") AS JSON)) as leaderboard 
	FROM 
	(SELECT u.id, u.name, u.nickname, u.lbmeta, ru.score, u.country, ru.room_id 
	FROM users u, rooms_users ru WHERE u.id=ru.user_id AND ru.room_id=(SELECT MAX(room_id) FROM rooms_users r WHERE r.user_id = ?) LIMIT 10)  l WHERE l.id != ?) as q,
	(SELECT score from rooms_users ru WHERE ru.user_id=? AND ru.room_id=(SELECT MAX(room_id) FROM rooms_users r WHERE r.user_id = ?)) as score,
	(SELECT COUNT(*) + 1 as rank FROM rooms_users j WHERE j.user_id=? AND j.room_id=(SELECT MAX(room_id) FROM rooms_users r WHERE r.user_id = ?) AND j.score > score) as rank`
)
