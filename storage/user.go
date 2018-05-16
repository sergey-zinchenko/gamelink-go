package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
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
func (u User) Data() (C.J, error) {
	var bytes []byte
	if u.dbs.mySQL == nil {
		return nil, errors.New("databases not initialized")
	}
	err := u.dbs.mySQL.QueryRow("SELECT `data` FROM `users` WHERE `id` = ?", u.ID()).Scan(&bytes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	var data C.J
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *User) txData(tx *sql.Tx) (C.J, error) {
	var bytes []byte
	err := tx.QueryRow("SELECT `data` FROM `users` u WHERE u.`id`=?", u.ID()).Scan(&bytes)
	if err != nil {
		return nil, err
	}
	var data C.J
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *User) txUpdate(data C.J, tx *sql.Tx) error {
	upd, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE `users` u SET u.`data`=? WHERE u.id=?", upd, u.ID())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) txDelete(tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM `users` WHERE `id`=?", u.ID())
	return err
}

func (u *User) logout() error {
	return nil
}

//Update - allow user update data
func (u *User) Update(data C.J) (C.J, error) {
	var transaction = func(upd C.J, tx *sql.Tx) (C.J, error) {
		data, err := u.txData(tx)
		if err != nil {
			return nil, err
		}
		for k, v := range upd {
			data[k] = v
		}
		err = u.txUpdate(data, tx)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	delete(data, "fb_id")
	delete(data, "vk_id")
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	data, err = transaction(data, tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, e
		}
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Delete - allow user delete data about him or delete account
func (u *User) Delete(fields []string) (C.J, error) {
	var transaction = func(fields []string, tx *sql.Tx) (C.J, error) {
		if len(fields) != 0 {
			data, err := u.txData(tx)
			if err != nil {
				return nil, err
			}
			for _, v := range fields {
				if v == "fb_id" || v == "vk_id" {
					continue
				}
				delete(data, v)
			}
			err = u.txUpdate(data, tx)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
		err := u.txDelete(tx)
		if err != nil {
			return nil, err
		}
		err = u.logout() // Redis Call - Delete tokens from Redis
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	updData, err := transaction(fields, tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, err
		}
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return updData, nil
}

// AddSocial - allow
func (u *User) AddSocial(token social.ThirdPartyToken) (C.J, error) {
	var transaction = func(ID social.ThirdPartyID, tx *sql.Tx) (C.J, error) {
		data, err := u.txData(tx)
		if err != nil {
			return nil, err
		}
		if _, ok := data[ID.Name()]; ok {
			return nil, graceful.BadRequestError{Message: "account already exist"}
		}
		data[ID.Name()] = ID.Value()
		err = u.txUpdate(data, tx)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("empty token")
	}
	id, _, _, err := token.UserInfo()
	if err != nil {
		return nil, err
	}
	updData, err := transaction(id, tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, err
		}
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return updData, nil
}

//TODO: Нужно создать обраюотчик для добавление аторизации по второй социалке для зарегистрированного пользователя
//Пользователь зарегистрированный через vk ожет захотеть добавить авторизацию через fb и наоборот
//http method GET for path /users/auth
//URL какого-то такого вида /users/auth?fb=sometoken
//само собой только для авторизованных пользователей
//через транзакции
//возможные ситуации такие: у пользователя уже задан токен такой социалки или есть другая заптсь о пользователе с такой социалкой - вернуть Bad Request.
