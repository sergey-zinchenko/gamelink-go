package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"gamelink-go/storage/queries"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type (
	//User - structure to work with user in our system. Developed to be passed through context of request.
	User struct {
		id  int64
		dbs *DBS
	}
	//Device - struct for device data
	Device struct {
		deviceID   string
		deviceType string
	}
)

//ID - returns user's id from database
func (u User) ID() int64 {
	return u.id
}

func (u *User) txCheck(userData social.ThirdPartyUser, tx *sql.Tx) (bool, error) {
	if userData.ID() == nil {
		return false, nil
	}
	var deletedFlag int
	queryString := fmt.Sprintf(queries.CheckUserQuery, userData.ID().Name())
	err := tx.QueryRow(queryString, userData.ID().Value()).Scan(&u.id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	q := fmt.Sprintf(queries.CheckFlag, userData.ID().Name())
	err = tx.QueryRow(q, userData.ID().Value()).Scan(&deletedFlag)
	if deletedFlag == 1 {
		err = graceful.ForbiddenError{Message: "user was deleted"}
		return false, err
	}
	return true, nil
}

func (u *User) txRegister(user social.ThirdPartyUser, tx *sql.Tx) error {
	b, err := json.Marshal(user)
	if err != nil {
		return err
	}
	res, err := tx.Exec(queries.RegisterUserQuery, b)
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
		if registered && (user.Country() != "" || user.BirthDate() != 0 || user.Sex() != "") {
			data := make(C.J)
			if user.Country() != "" {
				data["country"] = user.Country()
			}
			if user.BirthDate() != 0 {
				data["bdate"] = user.BirthDate()
			}
			if user.Sex() != "" {
				data["sex"] = user.Sex()
			}
			upd, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = tx.Exec(queries.UpdateSexCountryBdate, upd, u.ID(), u.ID())
			if err != nil {
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
	var str sql.NullString
	if u.dbs.mySQL == nil {
		return "", errors.New("databases not initialized")
	}
	err := u.dbs.mySQL.QueryRow(queries.GetExtraUserDataQuery, u.ID(), u.ID(), u.ID(), u.ID(), u.ID(), u.ID(), u.ID(), u.ID(), u.ID()).Scan(&str)
	if err != nil {
		logrus.Warn(err.Error())
		if err == sql.ErrNoRows {
			return "", graceful.NotFoundError{Message: "user not found"}
		}
		return "", err
	}
	if !str.Valid {
		return "", graceful.NotFoundError{Message: "user not found"}
	}
	return str.String, nil
}

func (u User) txData(tx *sql.Tx) (C.J, error) {
	var bytes []byte
	err := tx.QueryRow(queries.GetUserDataQuery, u.ID()).Scan(&bytes)
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
	_, err = tx.Exec(queries.UpdateUserDataQuery, upd, u.ID())
	if err != nil {
		switch v := err.(type) {
		case *mysql.MySQLError:
			if v.Number == mysqlKeyExist {
				return graceful.ConflictError{Message: "user with this social account already exist"}
			}
		default:
			return err
		}
	}
	return nil
}

func (u User) txDelete(tx *sql.Tx) error {
	_, err := tx.Exec(queries.DeleteAllSaves, u.ID())
	if err != nil {
		return err
	}
	_, err = tx.Exec(queries.DeleteUserQuery, u.ID())
	if err != nil {
		return err
	}
	_, err = tx.Exec(queries.DeleteUserFromFriends, u.ID(), u.ID())
	if err != nil {
		return err
	}

	return err
}

func (u User) txSyncFriends(friendsIds []social.ThirdPartyID, tx *sql.Tx) error {
	var err error
	vkStmt, err := tx.Prepare(fmt.Sprintf(queries.MakeFriendshipQuery, social.VkID))
	if err != nil {
		return err
	}
	defer vkStmt.Close()
	fbStmt, err := tx.Prepare(fmt.Sprintf(queries.MakeFriendshipQuery, social.FbID))
	if err != nil {
		return err
	}
	defer fbStmt.Close()
	for _, v := range friendsIds {
		switch v.(type) {
		case social.VkIdentifier:
			_, err = vkStmt.Exec(u.ID(), v.Value(), v.Value(), u.ID())
		case social.FbIdentifier:
			_, err = fbStmt.Exec(u.ID(), v.Value(), v.Value(), u.ID())
		}
		if err != nil {
			return err
		}
	}
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
	delete(data, "name")
	delete(data, "country")
	delete(data, "bdate")
	delete(data, "email")
	delete(data, "sex")
	if data["nickname"] != nil && data["nickname"] == "" {
		delete(data, "nickname")
	}
	meta := data["meta"].(map[string]interface{})
	if len(meta) == 0 {
		delete(data, "meta")
	}
	saves := data["saves"].([]interface{})
	if len(saves) == 0 {
		delete(data, "saves")
	}
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
				if v == "fb_id" || v == "vk_id" || v == "name" || v == "country" || v == "bdate" || v == "email" || v == "sex" {
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
		var isDummy bool
		if data["vk_id"] == nil && data["fb_id"] == nil {
			isDummy = true
		}
		if err != nil {
			return nil, err
		}

		//if _, ok := data[userData.ID().Name()]; ok {
		//	return nil, graceful.BadRequestError{Message: "account already exist"}
		//}
		data[userData.ID().Name()] = userData.ID().Value()
		if isDummy {
			data["name"] = userData.Name()
			data["sex"] = userData.Sex()
			data["email"] = userData.Email()
			data["country"] = userData.Country()
			data["bdate"] = userData.BirthDate()
		}
		err = u.txUpdate(data, tx)
		if err != nil {
			//Обработчик 409
			switch err.(type) {
			case graceful.ConflictError:
				err = u.ReloginWithUpdate(data, userData, tx)
				if err != nil {
					return nil, err
				}
				return data, nil
			}
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
		return nil, graceful.BadRequestError{Message: "empty token"}
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

//ReloginWithUpdate - delete dummy user from db then update old account with social and change its id to deleted dummy id
func (u User) ReloginWithUpdate(data C.J, thirdPartyUserData social.ThirdPartyUser, tx *sql.Tx) error {
	_, err := tx.Exec(queries.DeleteDummyUserFromDB, u.ID())
	if err != nil {
		return err
	}
	q := fmt.Sprintf(queries.UpdateRecoveryUserDataAndIDQuery, thirdPartyUserData.ID().Name())
	upd, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = tx.Exec(q, u.ID(), upd, thirdPartyUserData.ID().Value())
	if err != nil {
		return err
	}
	return nil
}

//AddDevice - add device id and type to device_ids table
func (u User) AddDevice(device *Device) error {
	_, err := u.dbs.mySQL.Exec(queries.AddDeviceID, u.ID(), device.deviceID, device.deviceType)
	if err != nil {
		return err
	}
	return nil
}

//DeleteDummyCreateNormalRedisToken - delete dummy user from redis
func (u User) DeleteDummyCreateNormalRedisToken(redisToken string) (string, error) {
	var err error
	var newToken string
	for ok := false; !ok; {
		newToken = C.RandStringBytes(40)
		authKey := AuthRedisKeyPref + newToken
		lifetime := time.Hour * 8
		ok, err = u.dbs.rc.SetNX(authKey, u.ID(), lifetime).Result()
		if err != nil {
			return "", err
		}
	}
	cmd := u.dbs.rc.Del(AuthRedisKeyPref + redisToken)
	if cmd.Err() != nil {
		logrus.Warn("redis delete token error", cmd.Err())
		return "", cmd.Err()
	}
	return newToken, nil
}
