package queries

const (
	//SelectLastTournament - query to select last tournament
	SelectLastTournament = `SELECT IFNULL((SELECT MAX(expired_time) FROM tournaments),0)`

	//CreateNewTournament - query to create new tournament
	CreateNewTournament = `INSERT INTO tournaments (expired_time) VALUES (?)`

	//CreateNewRoom - query to create new tournament room
	CreateNewRoom = `INSERT INTO rooms (expired_time) VALUES (?)`

	//JoinTournament - query to add user id in table users_tournaments to allow us to check if user already in tournament
	JoinTournament = `INSERT INTO users_tournaments (tournament_id, user_id) VALUES ((SELECT MAX(id) FROM tournaments), ?)`

	//GetCountUsersInRoomAndTournamentExpiredTime - query to get count of users in room to allow us to check max count users in room in current tournament
	GetCountUsersInRoomAndTournamentExpiredTime = `SELECT COUNT(user_id), expired_time FROM rooms_users WHERE room_id = (SELECT MAX(room_id) FROM rooms_users) `

	//JoinUserToExistRoom - query to join user in existed room
	JoinUserToExistRoom = `INSERT INTO rooms_users (room_id,expired_time, user_id) VALUES ((SELECT MAX(id) FROM rooms),(SELECT MAX(expired_time) FROM tournaments), ?)`

	//CreateNewRoomInCurrentTournament - query to create new room if there max users in last created room
	CreateNewRoomInCurrentTournament = `INSERT INTO rooms (expired_time) VALUES ((SELECT MAX(expired_time) FROM tournaments))`

	//JoinNewRoom - query to join user in created room
	JoinNewRoom = `INSERT INTO rooms_users (room_id,expired_time, user_id) VALUES ((SELECT MAX(id) FROM rooms),(SELECT MAX(expired_time) FROM tournaments), ?)`

	//UpdateUserTournamentScore - query to update user tournament score
	UpdateUserTournamentScore = `UPDATE rooms_users ru SET ru.score = ? WHERE user_id = ? AND expired_time > ?`
)
