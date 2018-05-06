package storage

import (
	"encoding/json"
	"errors"
)

type (
	//User - structure to work with user in our system. Developed to be passed through context of request.
	User struct {
		id  int64
		dbs *DBS
	}
)

//ID - returns user's id from database
func (u User) ID() int64 {
	return u.id
}

//Data - returns user's field data from database
func (u User) Data() (map[string]interface{}, error) {
	if u.dbs.mySQL == nil {
		return nil, errors.New("databases not initialized")
	}
	stmt, err := u.dbs.mySQL.Prepare("SELECT `data` FROM `users` WHERE `id` = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(u.ID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bytes []byte
	if !rows.Next() {
		return nil, errors.New("user not found")
	}
	err = rows.Scan(&bytes)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
