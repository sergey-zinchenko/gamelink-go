package storage

import (
	"database/sql"
	"fmt"

	"gamelink-go/proto_msg"
)

type (
	//QueryBuilder - struct fo work with params and build db query
	QueryBuilder struct {
		query, whereClause, message string
		offset                      int64
		params                      []interface{}
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
		q.WithClause(v)
	}
	return q
}

//Offset - set offset to qeryBuilder
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

//String - concatenate first query part, WHERE and params query part
func (q *QueryBuilder) String(offset int64) string {
	if q.query == "" {
		return ""
	}
	if q.whereClause == "" {
		return q.query + fmt.Sprintf(" LIMIT 1 OFFSET %d", offset)
		//return q.query
	}
	return fmt.Sprintf("%s WHERE %s", q.query, q.whereClause)
}

//QueryWithDB - execute query, scan result
func (q *QueryBuilder) QueryWithDB(sql *sql.DB, worker RowWorker) ([]interface{}, error) {
	rows, err := sql.Query(q.String(q.offset), q.params...)
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
