package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/proto_msg"
	"gamelink-go/storage/queries"
)

type (
	//QueryBuilder - struct fo work with params and build db query
	QueryBuilder struct {
		query, whereClause, message   string
		loggedInLess, loggedInGreater string
		offset                        int64
		params                        []interface{}
	}
	//UpdateBuilder - struct for work with params when update user data
	UpdateBuilder struct {
		ID        int64
		Data      []byte
		UpdParams []*proto_msg.UpdateCriteriaStruct
	}
	//ScanFunc - func for scan rows
	ScanFunc = func(...interface{}) error
	//RowWorker - use scanfunc, return query executing result
	RowWorker = func(ScanFunc) (interface{}, error)
)

//WithClause - func for make query part from criteria
func (q *QueryBuilder) WithClause(criteria *proto_msg.OneCriteriaStruct) *QueryBuilder {
	if q.whereClause != "" {
		q.whereClause += " AND "
	}
	q.whereClause += criteria.Cr.String() + " "
	switch criteria.Op {
	case proto_msg.OneCriteriaStruct_l:
		q.whereClause += "<= ?"
	case proto_msg.OneCriteriaStruct_e:
		q.whereClause += "= ?"
	case proto_msg.OneCriteriaStruct_g:
		q.whereClause += ">= ?"
	}
	q.params = append(q.params, criteria.Value)
	return q
}

//WithMultipleClause - loop from array of criterias
func (q *QueryBuilder) WithMultipleClause(criterias []*proto_msg.OneCriteriaStruct) *QueryBuilder {
	for _, v := range criterias {
		if v.Cr.String() == "message" {
			q.message = v.Value //Пишем в структуру сообщение на случай, если это запрос на отправку пуш сообщения
			continue
		}
		if v.Cr.String() == "logged_id" {
			if v.Op == proto_msg.OneCriteriaStruct_e || v.Op == proto_msg.OneCriteriaStruct_g {
				q.loggedInGreater = v.Value
			} else if v.Op == proto_msg.OneCriteriaStruct_l {
				q.loggedInLess = v.Value
			}
			continue
		}
		q.WithClause(v)
	}
	return q
}

//Offset - set offset to queryBuilder
func (q *QueryBuilder) Offset(offset int64) *QueryBuilder {
	q.offset = offset
	return q
}

//CountQuery - fist part of count query
func (q *QueryBuilder) CountQuery() *QueryBuilder {
	q.query = "SELECT COUNT(id) FROM users"
	return q
}

//SelectQuery - first part of select query
func (q *QueryBuilder) SelectQuery() *QueryBuilder {
	q.query = `SELECT id, vk_id, fb_id, name, email, sex, timestampdiff(YEAR, bdate, curdate()), country, date(created_at), deleted from users`
	return q
}

//DeleteQuery - first part of delete query
func (q *QueryBuilder) DeleteQuery() *QueryBuilder {
	q.query = `UPDATE users SET deleted=1`
	return q
}

//PushQuery - first part of query for find users who will recieve push message
func (q *QueryBuilder) PushQuery() *QueryBuilder {
	q.query = `**********************`
	return q
}

//GetData - get user data from db
func (q *QueryBuilder) GetData() *QueryBuilder {
	q.query = `SELECT id, data FROM users `
	return q
}

func (q *QueryBuilder) isGetDataQuery() bool {
	if q.query == `SELECT id, data FROM users ` {
		return true
	}
	return false
}

//Concat - concatenate first query part, WHERE and params query part
func (q *QueryBuilder) Concat(offset int64) string {
	if q.query == "" {
		return ""
	}
	if q.whereClause == "" {
		return q.query
	}
	query := fmt.Sprintf("%s WHERE %s", q.query, q.whereClause)
	if q.isGetDataQuery() {
		query += fmt.Sprintf(" LIMIT 100 OFFSET %d", offset)
	}
	return query
}

//QueryWithDB - execute query, scan result
func (q *QueryBuilder) QueryWithDB(sql *sql.DB, worker RowWorker) ([]interface{}, error) {
	rows, err := sql.Query(q.Concat(q.offset), q.params...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var res []interface{}
	for rows.Next() {
		oneres, err := worker(rows.Scan)
		if err != nil {
			return nil, err
		}
		res = append(res, oneres)
	}
	return res, nil
}

//Message - return query builder message
func (q *QueryBuilder) Message() string {
	return q.message
}

//Prepare - get user json, update fields, adn prepare to update it in db
func (u *UpdateBuilder) Prepare() (*UpdateBuilder, error) {
	var dataJSON C.J
	err := json.Unmarshal(u.Data, &dataJSON)
	if err != nil {
		return nil, err
	}
	for _, v := range u.UpdParams {
		if v.Uop == proto_msg.UpdateCriteriaStruct_set {
			dataJSON[v.Ucr.String()] = v.Value
		} else if v.Uop == proto_msg.UpdateCriteriaStruct_delete {
			delete(dataJSON, v.Ucr.String())
		}
	}
	u.Data, err = json.Marshal(dataJSON)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Update - execute db query
func (u *UpdateBuilder) Update(sql *sql.DB) error {
	_, err := sql.Exec(queries.AdminUpdateUserDataQuery, u.Data, u.ID)
	if err != nil {
		return err
	}
	return nil
}
