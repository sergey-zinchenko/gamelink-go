package queries

const (

	//CreateNewTournament - query to create new tournament
	CreateNewTournament = `INSERT INTO tournaments (users_in_room, registration_expired_time, tournament_expired_time) VALUES (?,?,?)`

	//CreateNewRoom - query to create new tournament room
	CreateNewRoom = `INSERT INTO rooms (tournament_id) VALUES ((SELECT MAX(id) from tournaments))`

	//JoinTournamentProc - call procedure
	JoinTournamentProc = `CALL gamelink.join_tournament(?, ?);`

	//UpdateUserTournamentScore - query to update user tournament score
	UpdateUserTournamentScore = `UPDATE rooms_users SET score = ? WHERE tournament_id = ? AND user_id = (SELECT id from users u WHERE u.id = ? AND u.deleted != 1) AND tournament_expired_time > ?`

	//GetRoomLeaderboard - query to get tournament leaderboard
	GetRoomLeaderboard = `SELECT CAST(CONCAT(
    '{"id":', i.id, ',',
    '"nickname":', JSON_QUOTE(IFNULL(i.nickname,i.name)), ',',
    '"score":', IFNULL(JSON_QUOTE(score),0), ',',
    '"rank":', rank, ',',
    IFNULL(CONCAT('"country":', JSON_QUOTE(i.country), ','),''),
    IFNULL(CONCAT('"meta":', i.meta, ','),''),
    '"leaderboard":', leaderboard,'}') AS JSON) as leaderboard 
           FROM (SELECT u.id, u.name, u.nickname, u.country,u.meta FROM users u WHERE u.id=?) as i,
					 (SELECT (CAST(IFNULL(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
																		  '"id":', 			l.id, ',',
                                                                          '"nickname":', 	JSON_QUOTE(IFNULL(l.nickname,l.name)), ',',
                                                                          '"score":', 		IFNULL(JSON_QUOTE(l.score), 0),
																		   IFNULL(CONCAT(',"country":', JSON_QUOTE(l.country)),''),
																		   IFNULL(CONCAT(',"meta":', l.meta),''),
                                                                          '}')), ']'), "[]") AS JSON)) as leaderboard 
	FROM 
	(SELECT u.id, u.name, u.nickname, u.meta, ru.score, u.country, ru.room_id 
	FROM users u, rooms_users ru WHERE u.id=ru.user_id AND ru.room_id=(SELECT room_id FROM rooms_users WHERE tournament_id = ? AND user_id =?) ORDER BY score LIMIT 10 ) l WHERE l.id != ? ) as q,
    (SELECT score FROM rooms_users WHERE tournament_id = ? AND user_id = ?) as score,
	(SELECT count(*)+1 as rank FROM rooms_users WHERE room_id=(SELECT room_id FROM rooms_users WHERE tournament_id = ? AND user_id = ?) AND score > IFNULL((SELECT score FROM rooms_users WHERE tournament_id = ? AND user_id = ?),0)) as rank`

	//GetAvailableTournaments - query to get all available tournaments
	GetAvailableTournaments = ` SELECT IFNULL(CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
									'"id":', t.id,
									',', '"registration_expired_time":', t.registration_expired_time,
									',', '"tournament_expired_time":', t.tournament_expired_time,
									'}')), ']') AS JSON),"[]") FROM tournaments t WHERE registration_expired_time > ?`

	//GetResults - query to get all user results from last 100 tournaments
	GetResults = ` SELECT IFNULL(CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{',
									'"id":', p.tournament_id,
								 	',', '"rank":',  p.rank,
									',', '"score":', JSON_QUOTE(p.score),
									'}')), ']') AS JSON),"[]") as results
FROM (SELECT t.tournament_id, t.score, (SELECT count(*)+1 as rank FROM rooms_users WHERE room_id=t.room_id AND score > t.score) as rank 
FROM (SELECT tournament_id, room_id, score FROM rooms_users t WHERE user_id = (SELECT id FROM users u WHERE u.id = ? AND u.deleted != 1) LIMIT 100) as t) as p `
)
