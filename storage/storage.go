package storage

import (
	"database/sql"
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
	dbs.mySQL, err = sql.Open("mysql", config.MysqlDsn)
	if err != nil {
		return
	}

	if err = dbs.mySQL.Ping(); err != nil {
		err = dbs.CreateDB()
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

//CreateDB - create schema and tables if not exist
func (dbs *DBS) CreateDB() (err error) {

	db, err := sql.Open("mysql", "admin:123@tcp(127.0.0.1:3306)/")
	if err != nil {
		fmt.Println("1")
		return
	}

	_, err = db.Exec(queries.CreateSchema)
	if err != nil {
		fmt.Println("2")
		return
	}

	_, err = db.Exec(queries.UseSchema)
	if err != nil {
		fmt.Println("3")
		return
	}
	_, err = db.Exec(queries.CreateTableUsers)
	if err != nil {
		fmt.Println("4")
		return
	}
	_, err = db.Exec(queries.CreateTableFriends)
	if err != nil {
		fmt.Println("5")
		return
	}
	_, err = db.Exec(queries.CreateTableSaves)
	if err != nil {
		fmt.Println("6")
		return
	}

	_, err = db.Exec(queries.CreateTableTournaments)
	if err != nil {
		fmt.Println("7")
		return
	}

	_, err = db.Exec(queries.CreateTableRooms)
	if err != nil {
		fmt.Println("8")
		return
	}

	_, err = db.Exec(queries.CreateTableRoomsUsers)
	if err != nil {
		fmt.Println("9")
		return
	}

	_, err = db.Exec(queries.CreateFunctionStartTournament)
	if err != nil {
		fmt.Println("10")
		return
	}

	_, err = db.Exec(queries.CreateFunctionJoinTournament)
	if err != nil {
		fmt.Println("11")
		return
	}

	for k := 1; k < NumOfLeaderBoards+1; k++ {
		viewCreationScript := fmt.Sprintf(queries.CreateLbView, k)
		_, err = db.Exec(viewCreationScript)
		if err != nil {
			fmt.Println("12")
			break
		}
	}
	dbs.Connect()
	return nil
}
