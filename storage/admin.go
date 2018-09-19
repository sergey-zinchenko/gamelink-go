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

//CountQuery - fist part of count query
func (q *QueryBuilder) CountQuery() *QueryBuilder {
	q.query = "SELECT COUNT(id) FROM users"
	return q
}

//SelectQuery - first part of select query
func (q *QueryBuilder) SelectQuery() *QueryBuilder {
	q.query = `SELECT (SELECT CAST(CONCAT( 	'{"id":'  , 	id,
								IFNULL(CONCAT(', "deleted":'  , 	deleted),""),
								IFNULL(CONCAT(',"vk_id":'    , 	JSON_QUOTE(vk_id)),""), 
                                IFNULL(CONCAT(',"fb_id":'    , 	JSON_QUOTE(fb_id)),""),
                                IFNULL(CONCAT(',"name":'  	 , 	JSON_QUOTE(name)),""),
								IFNULL(CONCAT(',"sex":'  	 , 	JSON_QUOTE(sex)),""),
								IFNULL(CONCAT(',"email":'    , 	JSON_QUOTE(email)),""),
                                IFNULL(CONCAT(',"bdate":'  , 	timestampdiff(YEAR, bdate, curdate())),""), 
                                IFNULL(CONCAT(',"country":'  , 	JSON_QUOTE(country)),""),
                                IFNULL(CONCAT(',"created_at":'  ,UNIX_TIMESTAMP(created_at)),""),
                                '}') AS JSON)) from users`
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

//String - concatenate first query part, WHERE and params query part
func (q *QueryBuilder) String() string {
	if q.query == "" {
		return ""
	}
	if q.whereClause == "" {
		return q.query
	}
	return fmt.Sprintf("%s WHERE %s", q.query, q.whereClause)
}

//QueryWithDB - execute query, scan result
func (q *QueryBuilder) QueryWithDB(sql *sql.DB, worker RowWorker) ([]interface{}, error) {
	rows, err := sql.Query(q.String(), q.params...)
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
