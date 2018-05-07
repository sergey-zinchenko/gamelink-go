package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
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

// UpdateData - update user data
func (u *User) UpdateData(userID int64, oldData map[string]interface{}, newData map[string]interface{}) error {
	delete(newData, "fb_id")
	delete(newData, "vk_id")
	for k, v := range newData {
		oldData[k] = v
	}
	var transaction = func(userID int64, Data *map[string]interface{}, tx *sql.Tx) error {
		stmt, err := tx.Prepare("UPDATE `users` SET `data`=? WHERE `id`=?")
		if err != nil {
			return err
		}
		b, err := json.Marshal(Data)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(b, userID)
		if err != nil {
			return err
		}
		return nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(userID, &oldData, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// DeleteData - delete user data
func (u *User) DeleteData(userID int64, queryValues url.Values, Data map[string]interface{}) error {
	var flag string
	var query string
	if len(queryValues) == 0 {
		query = "DELETE FROM `users` WHERE `id`=?"
		flag = "user_delete"
	} else {
		for _, v := range queryValues["data"] {
			if v == "fb_id" || v == "vk_id" {
				continue
			}
			delete(Data, v)
		}
		query = "UPDATE `users` SET `data`=? WHERE `id`=?"
		flag = "data_delete"
	}
	var transaction = func(userID int64, Data *map[string]interface{}, query string, tx *sql.Tx) error {
		stmt, err := tx.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		switch flag {
		case "user_delete":
			_, err = stmt.Exec(userID)
		case "data_delete":
			b, err := json.Marshal(Data)
			if err != nil {
				return err
			}
			_, err = stmt.Exec(b, userID)
		default:
			return errors.New("delete data error")
		}
		if err != nil {
			return err
		}
		return nil
	}
	tx, err := u.dbs.mySQL.Begin()
	if err != nil {
		return err
	}
	err = transaction(userID, &Data, query, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

//AddSocialAcc - allow user to add one more social media account
func (u *User) AddSocialAcc(userID int64, queryValues url.Values) error {
	var social string
	if vk, fb := queryValues["vk"], queryValues["fb"]; vk != nil && len(vk) == 1 && fb == nil {
		social = "vk_id"
	} else if fb != nil && len(fb) == 1 && vk == nil {
		social = "fb_id"
	}
	stmt, err := u.dbs.mySQL.Prepare("SELECT ? FROM `users` WHERE `id` = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(social, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var socialID string
	dbToken := rows.Next()
	if dbToken {
		err = rows.Scan(&socialID)
		if err != nil {
			return err
		}
	}
	if socialID != "" {
		err = errors.New("Bad request")
		return err
	}

	// Теперь, если id этой соцсети в базе пуст, нужно добавить адишник этой соцсети

	return nil
}

//TODO: Нужно создать обработчик для загрузки информации о зарегистрированном пользователе.
//http method post for path /users
//Можно грузить произвольный JSON
//В нем не должно быть полей vk_id, fb_id - их надо удалять из входящих данных
//Обработка должна быть не по тригеру (удали его из бд), а с использованием транзакции
//схему организации транзакции посмотри в ThirdPartyUser
//Сначала грузишь json c инфой из базы (data) потом объединяешь его с тем что получен в теле запроса и пишешь это братно через метод update

//TODO: Нужно создать обработчик для удаления полей из инфы о пользователе или всего пользователя
//http method delete for path /users
//если вызван DELETE /users/ с валидным  Authorization и в URL query нет параметров то нужно снести всего пользователя целиком - вызвать DELETE в базе по идшнику.
//если в URL Query содержит массив data - i.e DELETE /users?data=field1&data=field2 то нужно удалить поля с этими именами в json из поля data в базе.
//само собой нельзя грохать fb_id и vk_id
//так же как и в предыдущем кейсе делать через транзакции без тригеров

//TODO: Нужно создать обраюотчик для добавление аторизации по второй социалке для зарегистрированного пользователя
//Пользователь зарегистрированный через vk ожет захотеть добавить авторизацию через fb и наоборот
//http method GET for path /users/auth
//URL какого-то такого вида /users/auth?fb=sometoken
//само собой только для авторизованных пользователей
//через транзакции
//возможные ситуации такие: у пользователя уже задан токен такой социалки или есть другая заптсь о пользователе с такой социалкой - вернуть Bad Request.
