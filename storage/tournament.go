package storage

import (
	"errors"
	"time"
)

//tournamentLifeTime - tournament lifetime in seconds
//const tournamentLifeTime  = 28800
const tournamentLifeTime = 30

//const tournamentInterval  = 259200
const tournamentInterval = 60

//StartTournament - func to start new tournament
func (dbs DBS) StartTournament() error {
	var success int
	tournamentExpiredTime := time.Now().Unix() + tournamentLifeTime
	err := dbs.mySQL.QueryRow("SELECT start_tournament(?, ?)", tournamentExpiredTime, tournamentInterval).Scan(&success)
	if err != nil {
		return err
	}
	if success != 1 {
		err = errors.New("to early to start new tournament")
		return err
	}
	return nil
}

//Join - func to join user to tournament
func (u User) Join() error {
	var success int
	err := u.dbs.mySQL.QueryRow("SELECT join_tournament(?,?)", u.ID(), time.Now().Unix()).Scan(&success)
	if err != nil {
		return err
	}
	if success != 1 {
		err = errors.New("there is not active tournaments")
		return err
	}
	return nil
}
