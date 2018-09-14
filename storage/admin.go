package storage

import (
	"fmt"
)

type (
	//Admin - struct for work with db
	Admin struct {
		dbs *DBS
	}
)

//Admin - return struct for work with db
func (dbs DBS) Admin() *Admin {
	return &Admin{&dbs}
}

//Count - make count query to db
func (a Admin) Count(query string) (string, error) {
	var res string
	subquery := "SELECT COUNT(id) FROM users WHERE ?"
	fmt.Println(subquery + query)
	err := a.dbs.mySQL.QueryRow(subquery, query).Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

//Find - make find query to db
func (a Admin) Find(query string) (string, error) {
	return "", nil
}

//Delete - safe delete user in db (set 1 to deleted field)
func (a Admin) Delete(query string) (string, error) {
	subquery := "UPDATE users u SET u.deleted=1 WHERE ?"
	fmt.Println(query)
	res, err := a.dbs.mySQL.Exec(subquery, query)
	fmt.Println(res)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "can't delete user with this params", nil
	}
	return "success", nil
}
