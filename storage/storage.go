package storage

import (
	"database/sql"
	"gamelink-go/config"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql" //That blank import is required to add mysql driver to the app
)

//const (
//	// NumOfLeaderBoards - number of leaderboards
//	NumOfLeaderBoards = 3
//)

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
		dbs.mySQL.Close() //i dont know about correctness
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

////CreateDB - create schema and tables if not exist
//func (dbs *DBS) CreateDB() (err error) {
//	_, err = dbs.mySQL.Exec(queries.CreateSchema)
//	if err != nil {
//		return
//	}
//	_, err = dbs.mySQL.Exec(queries.CreateTableUsers)
//	if err != nil {
//		return
//	}
//	_, err = dbs.mySQL.Exec(queries.CreateTableFriends)
//	if err != nil {
//		return
//	}
//	_, err = dbs.mySQL.Exec(queries.CreateTableSaves)
//	if err != nil {
//		return
//	}
//	for k := 0; k < NumOfLeaderBoards; k++ {
//		_, err = dbs.mySQL.Exec(queries.CreateLbView, k)
//		if err != nil {
//			break
//			return
//		}
//	}
//	return nil
//}
