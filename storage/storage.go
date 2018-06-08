package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gamelink-go/config"
	"gamelink-go/storage/queries"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql" //That blank import is required to add mysql driver to the app
)

const (
	// NumOfLeaderBoards - number of leaderboards
	NumOfLeaderBoards = 3
)

type (
	//DBS - class to work with storage
	DBS struct {
		rc    *redis.Client
		mySQL *sql.DB
	}
)

//Connect - Connections to all databases will be established here.
func (dbs *DBS) Connect() (err error) {
	if dbs.mySQL, err = sql.Open("mysql", config.MysqlDsn); err != nil {
		return err
	}
	if err = dbs.mySQL.Ping(); err != nil {
		if err != nil {
			dbs.mySQL.Close() //i dont know about correctness
		}
		return
	}
	dbs.rc = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDb,
	})
	if _, err = dbs.rc.Ping().Result(); err != nil {
		dbs.rc.Close()    //i dont know about correctness
		dbs.mySQL.Close() //i dont know about correctness
		return
	}
	return
}

//CheckTables - create schema and tables if not exist
func (dbs *DBS) CheckTables() (err error) {
	if dbs.mySQL == nil {
		return errors.New("mysql database not connected")
	}
	if _, err = dbs.mySQL.Exec(queries.CreateTableUsers); err != nil {
		return
	}
	if _, err = dbs.mySQL.Exec(queries.CreateTableFriends); err != nil {
		return
	}

	if _, err = dbs.mySQL.Exec(queries.CreateTableSaves); err != nil {
		return
	}
	if _, err = dbs.mySQL.Exec(queries.CreateTableTournaments); err != nil {
		return
	}

	if _, err = dbs.mySQL.Exec(queries.CreateTableRooms); err != nil {
		return
	}

	if _, err = dbs.mySQL.Exec(queries.CreateTableRoomsUsers); err != nil {
		return
	}

	if _, err = dbs.mySQL.Exec(queries.CreateFunctionStartTournament); err != nil {
		return
	}

	if _, err = dbs.mySQL.Exec(queries.CreateFunctionJoinTournament); err != nil {
		return
	}

	for k := 1; k < NumOfLeaderBoards+1; k++ {
		viewCreationScript := fmt.Sprintf(queries.CreateLbView, k)
		if _, err = dbs.mySQL.Exec(viewCreationScript); err != nil {
			return
		}
	}
	return nil
}
