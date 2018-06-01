package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/storage/queries"
)

// Saves - return saves from db all or one by instance id
func (u User) Saves(saveID int) (string, error) {
	var str string
	var err error
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	if saveID == 0 {
		err = u.dbs.mySQL.QueryRow(queries.GetAllSavesQuery, u.ID()).Scan(&str)
	} else {
		err = u.dbs.mySQL.QueryRow(queries.GetSaveQuery, saveID, u.ID()).Scan(&str)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return "", graceful.NotFoundError{Message: "saves not found"}
		}
		return "", err
	}
	return str, nil
}

//txSaveData - returns save data in C.J format
func (u User) txSaveData(saveID int, tx *sql.Tx) (C.J, error) {
	var bytes []byte
	err := tx.QueryRow(queries.GetSaveDataQuery, saveID).Scan(&bytes)
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
func (u User) txUpdateSaveData(data C.J, saveID int, tx *sql.Tx) error {
	upd, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = tx.Exec(queries.UpdateSaveDataQuery, upd, saveID)
	if err != nil {
		return err
	}
	return nil
}

func (u User) txDeleteSave(saveID int, tx *sql.Tx) error {
	_, err := tx.Exec(queries.DeleteSaveQuery, saveID)
	return err
}

//UpdateSave - update save data in transaction, return updated data
func (u User) UpdateSave(data C.J, saveID int) (C.J, error) {
	var transaction = func(upd C.J, saveID int, tx *sql.Tx) (C.J, error) {
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
	var transaction = func(s []byte, tx *sql.Tx) error {
		_, err := tx.Exec(queries.CreateSaveQuery, s, u.ID())
		return err
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	s, err := json.Marshal(data)
	err = transaction(s, tx)
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

//DeleteSave - allow user save some data from save data or delete save
func (u User) DeleteSave(saveID int, fields []string) (C.J, error) {
	var transaction = func(saveID int, fields []string, tx *sql.Tx) (C.J, error) {
		if len(fields) != 0 {
			data, err := u.txSaveData(saveID, tx)
			if err != nil {
				return nil, err
			}
			for _, v := range fields {
				delete(data, v)
			}
			err = u.txUpdateSaveData(data, saveID, tx)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
		err := u.txDeleteSave(saveID, tx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return nil, err
	}
	updData, err := transaction(saveID, fields, tx)
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
