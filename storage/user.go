package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

func (u *User) txCheck(socialID social.ThirdPartyID, tx *sql.Tx) (bool, error) {
	queryString := fmt.Sprintf("SELECT `id` FROM `users` u WHERE u.`%s` = ?", socialID.Name())
	err := tx.QueryRow(queryString, socialID).Scan(&u.id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *User) txRegister(socialID social.ThirdPartyID, name string, tx *sql.Tx) error {
	b, err := json.Marshal(C.J{socialID.Name(): socialID, "name": name})
	if err != nil {
		return err
	}
	res, err := tx.Exec("INSERT INTO `users` (`data`) VALUES (?)", b)
	if err != nil {
		return err
	}
	u.id, err = res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

//LoginUsingThirdPartyToken - function to fill users id by third party token
func (u *User) LoginUsingThirdPartyToken(token social.ThirdPartyToken) error {
	var transaction = func(socialID social.ThirdPartyID, name string, friendIds []social.ThirdPartyID, tx *sql.Tx) error {
		registered, err := u.txCheck(socialID, tx)
		if err != nil {
			return err
		}
		if !registered {
			if err = u.txRegister(socialID, name, tx); err != nil {
				return err
			}
		}
		err = u.txSyncFriends(friendIds, tx)
		if err != nil {
			return err
		}
		return nil
	}
	socialID, name, friendsIds, err := token.UserInfo()
	if err != nil {
		return err
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(socialID, name, friendsIds, tx)
	if err != nil {
		u.id = 0
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
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

func (u User) txData(tx *sql.Tx) (C.J, error) {
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

func (u User) txUpdate(data C.J, tx *sql.Tx) error {
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

func (u User) txDelete(tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM `users` WHERE `id`=?", u.ID())
	return err
}

func (u User) txSyncFriends(friendsIds []social.ThirdPartyID, tx *sql.Tx) error {
	const queryString = "INSERT IGNORE INTO `friends` (`user_id1`, `user_id2`) SELECT GREATEST(ids.id1, ids.id2),   LEAST(ids.id1, ids.id2) FROM (SELECT ? as id1 , u2.id as id2 FROM (SELECT `id` FROM `users` u WHERE u.`%s` = ? ) u2) ids"
	var err error
	vkStmt, err := tx.Prepare(fmt.Sprintf(queryString, social.VkID))
	if err != nil {
		return err
	}
	defer vkStmt.Close()
	fbStmt, err := tx.Prepare(fmt.Sprintf(queryString, social.FbID))
	if err != nil {
		return err
	}
	defer fbStmt.Close()
	for _, v := range friendsIds {
		switch v.(type) {
		case social.VkIdentifier:
			_, err = vkStmt.Exec(u.ID(), v.Value())
		case social.FbIdentifier:
			_, err = fbStmt.Exec(u.ID(), v.Value())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (u User) logout() error {
	return nil
}

//Update - allow user update data
func (u User) Update(data C.J) (C.J, error) {
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
func (u User) Delete(fields []string) (C.J, error) {
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
func (u User) AddSocial(token social.ThirdPartyToken) (C.J, error) {
	var transaction = func(ID social.ThirdPartyID, friendIds []social.ThirdPartyID, tx *sql.Tx) (C.J, error) {
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
		err = u.txSyncFriends(friendIds, tx)
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
	id, _, friendIds, err := token.UserInfo()
	if err != nil {
		return nil, err
	}
	updData, err := transaction(id, friendIds, tx)
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
