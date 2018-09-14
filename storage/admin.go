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
	subquery := "SELECT COUNT(id) FROM users WHERE %s"
	q := fmt.Sprintf(subquery, query)
	stmt, err := a.dbs.mySQL.Prepare(q)
	if err != nil {
		return "", err
	}
	err = stmt.QueryRow().Scan(&res)
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
	subquery := "UPDATE users u SET u.deleted=1 WHERE %s"
	q := fmt.Sprintf(subquery, query)
	stmt, err := a.dbs.mySQL.Prepare(q)
	if err != nil {
		return "", err
	}
	res, err := stmt.Exec()
	fmt.Println(res)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "can't delete user with this params", nil
	}
	return "success", nil
}
