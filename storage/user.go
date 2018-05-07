package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	for ok := range oldData {
		for nk := range newData {
			if nk == "fb_id" || nk == "vk_id" {
				continue
			}
			if nk == ok {
				oldData[ok] = newData[nk]
			}
			oldData[nk] = newData[nk]
		}
	}
	fmt.Println(oldData)
	var transaction = func(userID int64, Data *map[string]interface{}, tx *sql.Tx) error {
		stmt, err := tx.Prepare("UPDATE `users` SET `data`=(?) WHERE `id`=(?)")
		if err != nil {
			return err
		}
		b, err := json.Marshal(Data)
		fmt.Println(b)
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
