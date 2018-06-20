package storage

import (
	"database/sql"
	"fmt"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
	"github.com/go-sql-driver/mysql"
	"time"
)

//tournamentLifeTime - tournament lifetime in seconds

const (
	mysqlKeyExist = 1062
)

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
	var expiredTime, countUsersInRoom, maxUsersInRoom int64
	var transaction = func(userID int64, tx *sql.Tx) error {
		_, err := tx.Exec(queries.JoinTournament, tournamentID, userID)
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
		err = tx.QueryRow(queries.GetCountUsersInRoomAndTournamentExpiredTime, tournamentID, tournamentID, tournamentID).Scan(&expiredTime, &countUsersInRoom, &maxUsersInRoom)
		if err != nil {
			return err
		}
		if expiredTime < time.Now().Unix() {
			fmt.Println(expiredTime)
			fmt.Println(time.Now().Unix())
			return graceful.ForbiddenError{Message: "registration time have been expired"}
		}
		if countUsersInRoom < maxUsersInRoom {
			_, err = tx.Exec(queries.JoinUserToRoom, tournamentID, tournamentID, userID)
			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec(queries.CreateNewRoomInCurrentTournament, tournamentID)
			if err != nil {
				return err
			}
			_, err = tx.Exec(queries.JoinUserToRoom, tournamentID, tournamentID, userID)
		}
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
	_, err := u.dbs.mySQL.Exec(queries.UpdateUserTournamentScore, score, tournamentID, u.ID())
	if err != nil {
		return err
	}
	return nil
}

//GetLeaderboard - method to get leaderbord from tournament room
func (u User) GetLeaderboard(tournamentID int) (string, error) {
	var result string
	err := u.dbs.mySQL.QueryRow(queries.GetRoomLeaderboard, u.ID(), tournamentID, u.ID(), u.ID(), tournamentID, u.ID(), tournamentID, u.ID(), tournamentID, u.ID()).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}
