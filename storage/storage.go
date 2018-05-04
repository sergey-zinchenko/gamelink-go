package storage

import (
	"database/sql"
	"gamelink-go/config"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql" //That blank import is required to add mysql driver to the app
)

type (
	//DBS - class to work with storage
	DBS struct {
		rc    *redis.Client
		mySQL *sql.DB
	}
)

//NewDBS - DBS constructor. Connections to all databases will be established here.
func NewDBS() (*DBS, error) {
	mySQL, err := sql.Open("mysql", config.MysqlDsn)
	if err != nil {
		return nil, err
	}
	if err = mySQL.Ping(); err != nil {
		mySQL.Close() //i dont know about correctness
		return nil, err
	}
	rc := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDb,
	})
	if _, err = rc.Ping().Result(); err != nil {
		rc.Close()    //i dont know about correctness
		mySQL.Close() //i dont know about correctness
		return nil, err
	}
	return &DBS{rc, mySQL}, nil
}
