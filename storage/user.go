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

func (u *User) txCheck(userData social.ThirdPartyUser, tx *sql.Tx) (bool, error) {
	queryString := fmt.Sprintf("SELECT `id` FROM `users` u WHERE u.`%s` = ?", userData.ID().Name())
	err := tx.QueryRow(queryString, userData.ID().Value()).Scan(&u.id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *User) txRegister(user social.ThirdPartyUser, tx *sql.Tx) error {
	b, err := json.Marshal(user)
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
	var transaction = func(user social.ThirdPartyUser, tx *sql.Tx) error {
		registered, err := u.txCheck(user, tx)
		if err != nil {
			return err
		}
		if !registered {
			if err = u.txRegister(user, tx); err != nil {
				return err
			}
		}
		err = u.txSyncFriends(user.Friends(), tx)
		if err != nil {
			return err
		}
		return nil
	}
	userData, err := token.UserInfo()
	if err != nil {
		return err
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(userData, tx)
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

//DataString - returns user's field data from database as text
func (u User) DataString() (string, error) {
	var str string
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	err := u.dbs.mySQL.QueryRow("SELECT IFNULL((SELECT JSON_INSERT(u.`data`, '$.friends', fj.`friends`) FROM `users` u, "+
		"(SELECT "+
		"CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{', "+
		"'\"id\":',		b.`id`, "+
		"',', '\"name\":',		JSON_QUOTE(b.`name`),"+
		"'}')),']') AS JSON"+
		") "+
		"AS `friends` "+
		"FROM "+
		"(SELECT u.`id`, u.`name`,f.user_id2 as g FROM `friends` f,`users` u WHERE `user_id2` = ? AND f.user_id1 = u.id"+
		" UNION "+
		"SELECT u.`id`, u.`name`, f.user_id1 as g FROM `friends` f, `users` u WHERE `user_id1` = ? AND f.user_id2 = u.id) b "+
		"GROUP BY b.g) fj "+
		"WHERE u.`id` = ?), q.`data`) `data` FROM `users` q WHERE q.`id`=?", u.ID(), u.ID(), u.ID(), u.ID()).Scan(&str)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("user not found")
		}
		return "", err
	}
	return str, nil
}

//Data - returns user's field data from database
//TODO вот это нам не нужно, если только где-то не понадобится вызов Data по коду см.выше реализацию
func (u User) Data() (C.J, error) {
	var bytes []byte
	if u.dbs.mySQL == nil {
		return nil, errors.New("databases not initialized")
	}
	err := u.dbs.mySQL.QueryRow("SELECT JSON_INSERT(u.`data`, '$.friends', fj.friends) from `users` u, "+
		"(SELECT "+
		"CAST(CONCAT('[',GROUP_CONCAT(DISTINCT CONCAT('{', "+
		"'\"id\":',		b.`id`, "+
		"',', '\"name\":',		JSON_QUOTE(b.`name`),"+
		"'}')),']') AS JSON"+
		") "+
		"AS `friends` "+
		"FROM "+
		"(SELECT u.`id`, u.`name`,f.user_id2 as g FROM `friends` f,`users` u WHERE `user_id2` = ? AND f.user_id1 = u.id"+
		" UNION "+
		"SELECT u.`id`, u.`name`, f.user_id1 as g FROM `friends` f, `users` u WHERE `user_id1` = ? AND f.user_id2 = u.id) b "+
		"GROUP BY b.g) fj "+
		"WHERE u.`id` = ?", u.ID(), u.ID(), u.ID()).Scan(&bytes)
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
	var transaction = func(userData social.ThirdPartyUser, tx *sql.Tx) (C.J, error) {
		data, err := u.txData(tx)
		if err != nil {
			return nil, err
		}
		if _, ok := data[userData.ID().Name()]; ok {
			return nil, graceful.BadRequestError{Message: "account already exist"}
		}
		data[userData.ID().Name()] = userData.ID().Value()
		err = u.txUpdate(data, tx)
		if err != nil {
			return nil, err
		}
		err = u.txSyncFriends(userData.Friends(), tx)
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
	userData, err := token.UserInfo()
	if err != nil {
		return nil, err
	}
	updData, err := transaction(userData, tx)
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

// Saves - return saves from db all or one by instance id
func (u User) Saves(saveID []string) (string, error) {
	var str string
	var err error
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	if len(saveID) == 0 {
		err = u.dbs.mySQL.QueryRow("SELECT CAST(CONCAT('[', GROUP_CONCAT(DISTINCT CONCAT('{', '\"id\":',		s.`id`, ',', '\"name\":',	JSON_QUOTE(s.`name`), '}')),']') AS JSON) FROM `saves` s  WHERE s.`user_id` = ? GROUP BY s.`user_id`", u.ID()).Scan(&str)
	} else {
		err = u.dbs.mySQL.QueryRow("SELECT JSON_OBJECT('id', s.`id`, `name`, s.`name`) FROM `saves` s WHERE s.`id` = ? AND s.`user_id`=?", saveID[0], u.ID()).Scan(&str)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("instances not found")
		}
		return "", err
	}
	return str, nil
}

//txSaveData - returns save data in C.J format
func (u User) txSaveData(saveID string, tx *sql.Tx) (C.J, error) {
	var bytes []byte
	err := tx.QueryRow("SELECT `data` FROM `saves` s WHERE s.`id`=?", saveID).Scan(&bytes)
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

//txUpdateSaveData - update save data in db
func (u User) txUpdateSaveData(data C.J, saveID string, tx *sql.Tx) error {
	upd, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE `saves` s SET s.`data`=? WHERE s.id=?", upd, saveID)
	if err != nil {
		return err
	}
	return nil
}

//UpdateSave - update save data in transaction, return updated data
func (u User) UpdateSave(data C.J, saveID string) (C.J, error) {
	var transaction = func(upd C.J, saveID string, tx *sql.Tx) (C.J, error) {
		data, err := u.txSaveData(saveID, tx)
		if err != nil {
			return nil, err
		}
		for k, v := range upd {
			data[k] = v
		}
		err = u.txUpdateSaveData(data, saveID, tx)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	data, err = transaction(data, saveID, tx)
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

//CreateSave - create new save instance in db
func (u User) CreateSave(data C.J) (C.J, error) {
	var transaction = func(data C.J, tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO `saves` s SET s.`data` = ?, s.`user_id = ?` ", data, u.ID())
		if err != nil {
			return err
		}
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	err = transaction(data, tx)
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
	return data, nil
}
