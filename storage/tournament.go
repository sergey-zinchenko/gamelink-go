package storage

import (
	"database/sql"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
	"github.com/go-sql-driver/mysql"
	"time"
)

type (
	//Tournament - structure to work with tournament in our system. Developed to be passed through context of request.
	Tournament struct {
		id  int64
		dbs *DBS
	}
)

const (
	mysqlKeyExist = 1062
)

//TID - returns tournament id from database
func (t Tournament) TID() int64 {
	return t.id
}

//StartTournament - func to start new tournament
func (dbs DBS) StartTournament(usersInRoom int64, tournamentDuration int64, registrationDuration int64) error {
	tournamentExpiredTime := time.Now().Unix() + tournamentDuration
	registrationExpiredTime := time.Now().Unix() + registrationDuration
	var transaction = func(usersInRoom int64, registrationExpiredTime int64, tournamentExpiredTime int64, tx *sql.Tx) error {
		_, err := tx.Exec(queries.CreateNewTournament, usersInRoom, registrationExpiredTime, tournamentExpiredTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(queries.CreateNewRoom)
		if err != nil {
			return err
		}
		return nil
	}
	tx, err := dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(usersInRoom, registrationExpiredTime, tournamentExpiredTime, tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

//Join - func to join user to tournament
func (u User) Join(tournamentID int) error {
	var registrationExpiredTime, tournamentExpiredTime, countUsersInRoom, maxUsersInRoom int64
	var transaction = func(userID int64, tx *sql.Tx) error {
		_, err := tx.Exec(queries.LockTableRoomsUsers)
		if err != nil {
			return err
		}
		result, err := tx.Exec(queries.JoinTournament, tournamentID, userID)
		if err != nil {
			switch v := err.(type) {
			case *mysql.MySQLError:
				if v.Number == mysqlKeyExist {
					return graceful.ForbiddenError{Message: "you have been already registered in tournament"}
				}
			default:
				return err
			}
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return graceful.NotFoundError{Message: "user or tournament doesn't found"}
		}

		err = tx.QueryRow(queries.GetCountUsersInRoomAndTournamentExpiredTime, tournamentID, tournamentID, tournamentID).Scan(&registrationExpiredTime, &tournamentExpiredTime, &countUsersInRoom, &maxUsersInRoom)
		if err != nil {
			return err
		}
		if registrationExpiredTime < time.Now().Unix() {
			return graceful.ForbiddenError{Message: "registration time have been expired"}
		}
		if countUsersInRoom < maxUsersInRoom {
			_, err = tx.Exec(queries.JoinUserToRoom, tournamentID, tournamentID, userID, tournamentExpiredTime)
			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec(queries.CreateNewRoomInCurrentTournament, tournamentID)
			if err != nil {
				return err
			}
			_, err = tx.Exec(queries.JoinUserToRoom, tournamentID, tournamentID, userID, tournamentExpiredTime)
		}
		if err != nil {
			return err
		}
		_, err = tx.Exec(queries.UnlockTableRoomsUsers)
		if err != nil {
			return err
		}
		return nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	userID := u.ID()
	err = transaction(userID, tx)
	if err != nil {
		_, err = tx.Exec(queries.UnlockTableRoomsUsers)
		if err != nil {
			return err
		}
		if e := tx.Rollback(); e != nil {
			return e
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

//UpdateTournamentScore - method to update user score
func (u User) UpdateTournamentScore(tournamentID int, score int) error {
	result, err := u.dbs.mySQL.Exec(queries.UpdateUserTournamentScore, score, tournamentID, u.ID(), time.Now().Unix())
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return graceful.NotFoundError{Message: "can't update score"}
	}
	return nil
}

//GetLeaderboard - method to get leaderbord from tournament room
func (u User) GetLeaderboard(tournamentID int) (string, error) {
	var result string
	var flag int
	err := u.dbs.mySQL.QueryRow(queries.IternalCheckFlag, u.ID()).Scan(&flag)
	if err != nil {
		return "", err
	}
	if flag == 1 {
		err = graceful.ForbiddenError{Message: "request for deleted user"}
		return "", err
	}
	err = u.dbs.mySQL.QueryRow(queries.GetRoomLeaderboard, u.ID(), tournamentID, u.ID(), u.ID(), tournamentID, u.ID(), tournamentID, u.ID(), tournamentID, u.ID()).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

//GetTournaments - method to Get Available Tournaments
func (u User) GetTournaments() (string, error) {
	var result string
	var flag int
	err := u.dbs.mySQL.QueryRow(queries.IternalCheckFlag, u.ID()).Scan(&flag)
	if err != nil {
		return "", err
	}
	if flag == 1 {
		err = graceful.ForbiddenError{Message: "request for deleted user"}
		return "", err
	}
	err = u.dbs.mySQL.QueryRow(queries.GetAvailableTournaments, time.Now().Unix()).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

//GetResults - method to get user results from last 100 tournaments
func (u User) GetResults() (string, error) {
	var result string
	var flag int
	err := u.dbs.mySQL.QueryRow(queries.IternalCheckFlag, u.ID()).Scan(&flag)
	if err != nil {
		return "", err
	}
	if flag == 1 {
		err = graceful.ForbiddenError{Message: "request for deleted user"}
		return "", err
	}
	err = u.dbs.mySQL.QueryRow(queries.GetResults, u.ID()).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}
