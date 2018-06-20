package queries

const (

	//CreateNewTournament - query to create new tournament
	CreateNewTournament = `INSERT INTO tournaments (users_in_room, registration_expired_time, tournament_expired_time) VALUES (?,?,?)`

	//CreateNewRoom - query to create new tournament room
	CreateNewRoom = `INSERT INTO rooms (tournament_id) VALUES ((SELECT MAX(id) from tournaments))`

	//JoinTournament - query to add user id in table users_tournaments to allow us to check if user already in tournament
	JoinTournament = `INSERT INTO users_tournaments (tournament_id, user_id) VALUES (?, ?)`

	//GetCountUsersInRoomAndTournamentExpiredTime - query to get count of users in room to allow us to check max count users in room in current tournament
	GetCountUsersInRoomAndTournamentExpiredTime = `SELECT t.registration_expired_time, t.tournament_expired_time, c.users_count, d.users_in_room  FROM 
		(SELECT registration_expired_time, tournament_expired_time FROM tournaments WHERE id = ?) as t,
		(SELECT IFNULL(count(user_id),0) as users_count FROM rooms_users WHERE room_id = (SELECT MAX(room_id) FROM rooms_users WHERE tournament_id = ?)) as c,
		(SELECT users_in_room FROM tournaments WHERE id=?) as d`

	//JoinUserToRoom - query to join user in room
	JoinUserToRoom = `INSERT INTO rooms_users (room_id,tournament_id, user_id, tournament_expired_time) VALUES ((SELECT MAX(id) FROM rooms WHERE tournament_id=?),?, ?, ?)`

	//CreateNewRoomInCurrentTournament - query to create new room if there max users in last created room
	CreateNewRoomInCurrentTournament = `INSERT INTO rooms (tournament_id) VALUES (?)`

	//UpdateUserTournamentScore - query to update user tournament score
	UpdateUserTournamentScore = `UPDATE rooms_users SET score = ? WHERE tournament_id = ? AND user_id = ? AND tournament_expired_time > ?`

	//GetRoomLeaderboard - query to get tournament leaderboard
	GetRoomLeaderboard = `SELECT CAST(CONCAT(
    '{"id":', i.id, ',',
    '"nickname":', JSON_QUOTE(IFNULL(i.nickname,i.name)), ',',
    '"score":', IFNULL(score,0), ',',
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
	FROM users u, rooms_users ru WHERE u.id=ru.user_id AND ru.room_id=(SELECT room_id FROM rooms_users WHERE tournament_id = ? AND user_id =?) ORDER BY score LIMIT 10 ) l WHERE l.id != ? ) as q,
    (SELECT score FROM rooms_users WHERE tournament_id = ? AND user_id = ?) as score,
	(SELECT count(*)+1 as rank FROM rooms_users WHERE room_id=(SELECT room_id FROM rooms_users WHERE tournament_id = ? AND user_id = ?) AND score > (SELECT score FROM rooms_users WHERE tournament_id = ? AND user_id = ?)) as rank`

	//GetAvailableTournaments - query to get all available tournaments
	GetAvailableTournaments = ` SELECT IFNULL(CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
									'"id":', t.id,
									',', '"registration_expired_time":', t.registration_expired_time,
									'}')), ']') AS JSON),"") FROM tournaments t WHERE registration_expired_time > ?`
)
