package storage

import (
	"database/sql"
	"errors"
	C "gamelink-go/common"
	"github.com/go-sql-driver/mysql"
	"time"
	"gamelink-go/storage/queries"
)

//tournamentLifeTime - tournament lifetime in seconds

const (
	//const tournamentLifeTime  = 28800
	tournamentLifeTime = 60

	//const tournamentInterval  = 259200
	tournamentInterval = 180

	//usersInRoom = 200
	usersInRoom = 2

	mysqlKeyExist = 1062
)

//StartTournament - func to start new tournament
func (dbs DBS) StartTournament() error {
	var lastExpiredTournamentTime int64
	tournamentExpiredTime := time.Now().Unix() + tournamentLifeTime
	var transaction = func(expiredTime int64, tx *sql.Tx) error {
		err := tx.QueryRow(queries.SelectLastTournament).Scan(&lastExpiredTournamentTime)
		if err != nil {
			return err
		}
		if (time.Now().Unix() - lastExpiredTournamentTime) < tournamentInterval {
			err = errors.New("to early to start new tournament")
			return err
		}
		_, err = tx.Exec(queries.CreateNewTournament, expiredTime)
		if err != nil {
			return err
		}
		_, err = tx.Exec(queries.CreateNewRoom, expiredTime)
		return nil
	}
	tx, err := dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(tournamentExpiredTime, tx)
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
func (u User) Join() error {
	var countUsersInRoom, expiredTime int64
	var transaction = func(userID int64, tx *sql.Tx) error {
		_, err := tx.Exec(queries.JoinTournament, userID)
		if err != nil {
			if err.(*mysql.MySQLError).Number == mysqlKeyExist {
				return errors.New("you have been already registered in tournament")
			}
			return err
		}
		err = tx.QueryRow(queries.GetCountUsersInRoomAndTournamentExpiredTime).Scan(&countUsersInRoom, &expiredTime)
		if err != nil {
			return err
		}
		if expiredTime < time.Now().Unix() {
			return errors.New("there is no active tournaments")
		}
		if countUsersInRoom % usersInRoom != 1 {
			_, err = tx.Exec(queries.JoinUserToExistRoom, userID)
			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec(queries.CreateNewRoomInCurrentTournament)
			if err != nil {
				return err
			}
			_, err = tx.Exec(queries.JoinNewRoom, userID)
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
func (u User) UpdateTournamentScore(data C.J) error {
	score := data["score"]
	_, err := u.dbs.mySQL.Exec(queries.UpdateUserTournamentScore, score, u.ID(), time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

//GetLeaderboard - method to get leaderbord from tournament room
func (u User) GetLeaderboard() (string, error) {
	//var result string
	//var err error
	//if u.dbs.mySQL == nil {
	//	return "", errors.New("databases not initialized")
	//}
	return "", nil
}
