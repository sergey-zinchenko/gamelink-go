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
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS gamelink DEFAULT CHARACTER SET utf8 ")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE gamelink")
	if err != nil {
		return
	}
	_, err = db.Exec(queries.CreateTableUsers)
	if err != nil {
		return
	}
	_, err = db.Exec(queries.CreateTableFriends)
	if err != nil {
		return
	}
	_, err = db.Exec(queries.CreateTableSaves)
	if err != nil {
		return
	}
	fmt.Println("asd")
	for k := 1; k < NumOfLeaderBoards+1; k++ {
		viewCreationScript := fmt.Sprintf(queries.CreateLbView, k)
		_, err = db.Exec(viewCreationScript)
		if err != nil {
			break
		}
	}
	dbs.Connect()
	return nil
}
