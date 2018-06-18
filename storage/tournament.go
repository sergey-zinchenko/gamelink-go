package storage

import (
	"database/sql"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
	"github.com/go-sql-driver/mysql"
	"time"
)

//tournamentLifeTime - tournament lifetime in seconds

const (
	//const tournamentLifeTime  = 28800
	tournamentLifeTime = 540

	//const tournamentInterval  = 72*time.Hour
	tournamentInterval = 600

	//usersInRoom = 200
	usersInRoom = 4

	mysqlKeyExist = 1062
)

//StartTournament - func to start new tournament
func (dbs DBS) StartTournament() error {
	var lastExpiredTournamentTime int64
	tournamentExpiredTime := time.Now().Unix() + tournamentLifeTime
	var transaction = func(expiredTime int64, tx *sql.Tx) error {
		err := tx.QueryRow(queries.SelectMaxExpiredTime).Scan(&lastExpiredTournamentTime)
		if err != nil {
			return err
		}
		if time.Since(time.Unix(lastExpiredTournamentTime, 0)) < tournamentInterval {
			err = graceful.ForbiddenError{Message: "to early to start new tournament"}
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
			switch v := err.(type) {
			case *mysql.MySQLError:
				if v.Number == mysqlKeyExist {
					return graceful.ForbiddenError{Message: "you have been already registered in tournament"}
				}
			default:
				return err
			}
		}
		err = tx.QueryRow(queries.GetCountUsersInRoomAndTournamentExpiredTime).Scan(&expiredTime, &countUsersInRoom)
		if err != nil {
			return err
		}
		if expiredTime < time.Now().Unix() {
			fmt.Println(expiredTime)
			fmt.Println(time.Now().Unix())
			return graceful.ForbiddenError{Message: "there is no active tournaments"}
		}
		if countUsersInRoom < usersInRoom {
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
	_, err := u.dbs.mySQL.Exec(queries.UpdateUserTournamentScore, score, time.Now().Unix(), u.ID())
	if err != nil {
		return err
	}
	return nil
}

//GetLeaderboard - method to get leaderbord from tournament room
func (u User) GetLeaderboard() (string, error) {
	var result string
	var roomID, score, expiredTime int64
	err := u.dbs.mySQL.QueryRow(queries.GetRoomScoreExpiredTime, u.ID()).Scan(&roomID, &score, &expiredTime)
	if err != nil {
		fmt.Println("sad")
		return "", err
	}
	err = u.dbs.mySQL.QueryRow(queries.GetRoomLeaderboard, score, u.ID(), roomID, u.ID(), roomID, expiredTime, score).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}
