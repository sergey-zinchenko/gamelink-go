package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/proto_msg"
)

const (
	//QueryCount - const for count query
	QueryCount QueryKind = iota
	//QuerySelect - const for Select query
	QuerySelect
	//QueryUpdate - const for update query
	QueryUpdate
	//QueryDelete - const for delete query
	QueryDelete

	limit = 1
)

type (
	//QueryBuilder - struct fo work with params and build db query
	QueryBuilder struct {
		kind        QueryKind
		whereClause string
		params      []interface{}
		updParams   []*proto_msg.UpdateCriteriaStruct
	}
	//ScanFunc - func for scan rows
	ScanFunc = func(...interface{}) error
	//RowWorker - use scanfunc, return query executing result
	RowWorker = func(ScanFunc) (interface{}, error)
	//QueryKind - num of kind db query
	QueryKind = int64
)

//WithClause - func for make query part from criteria
func (q *QueryBuilder) WithClause(criteria *proto_msg.OneCriteriaStruct) {
	if q.whereClause != "" {
		q.whereClause += " AND "
	}
	if criteria.Cr == proto_msg.OneCriteriaStruct_updated_at {
		q.whereClause += "unix_timestamp(" + criteria.Cr.String() + ") "
	} else {
		q.whereClause += criteria.Cr.String() + " "
	}
	switch criteria.Op {
	case proto_msg.OneCriteriaStruct_l:
		q.whereClause += "<= ?"
	case proto_msg.OneCriteriaStruct_e:
		q.whereClause += "= ?"
	case proto_msg.OneCriteriaStruct_g:
		q.whereClause += ">= ?"
	}
	q.params = append(q.params, criteria.Value)
}

//WithMultipleClause - loop from array of criterias
func (q QueryBuilder) WithMultipleClause(criterias []*proto_msg.OneCriteriaStruct) QueryBuilder {
	for _, v := range criterias {
		q.WithClause(v)
	}
	return q
}

//WithData - add update params to UpdateBuilder
func (q QueryBuilder) WithData(criterias []*proto_msg.UpdateCriteriaStruct) QueryBuilder {
	q.updParams = criterias
	return q
}

//CountQuery - fist part of count query
func (q QueryBuilder) CountQuery() QueryBuilder {
	q.kind = QueryCount
	return q
}

//SelectQuery - first part of select query
func (q QueryBuilder) SelectQuery() QueryBuilder {
	q.kind = QuerySelect
	return q
}

//UpdateQuery - query for update command
func (q QueryBuilder) UpdateQuery() QueryBuilder {
	q.kind = QueryUpdate
	return q
}

//DeleteQuery - first part of delete query
func (q QueryBuilder) DeleteQuery() QueryBuilder {
	q.kind = QueryDelete
	return q
}

//QueryWithDB - execute query, scan result
func (q QueryBuilder) QueryWithDB(mysql *sql.DB, worker RowWorker) ([]interface{}, error) {
	switch q.kind {
	case QuerySelect, QueryCount, QueryDelete:
		var query string
		switch q.kind {
		case QueryCount:
			query = `SELECT COUNT(id) FROM users`
		case QuerySelect:
			query = `SELECT id, vk_id, fb_id, name, email, sex, timestampdiff(YEAR, bdate, curdate()), country, date(created_at), deleted from users`
		case QueryDelete:
			query = `UPDATE users SET deleted=1`
		}
		if q.whereClause != "" {
			query = fmt.Sprintf("%s WHERE %s", query, q.whereClause)
		}
		rows, err := mysql.Query(query, q.params...)
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
	case QueryUpdate:
		tx, err := mysql.Begin()
		if err != nil {
			return nil, err
		}
		err = updateTransaction(q.whereClause, q.updParams, q.params, tx)
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
	}
	return nil, nil
}

//Prepare - get user json, update fields, adn prepare to update it in db
func prepareDataToUpdateInDb(rowData []byte, updateParams []*proto_msg.UpdateCriteriaStruct) ([]byte, error) {
	var dataJSON C.J
	err := json.Unmarshal(rowData, &dataJSON)
	if err != nil {
		return nil, err
	}
	for _, v := range updateParams {
		if v.Uop == proto_msg.UpdateCriteriaStruct_set {
			dataJSON[v.Ucr.String()] = v.Value
		} else if v.Uop == proto_msg.UpdateCriteriaStruct_delete {
			delete(dataJSON, v.Ucr.String())
		}
	}
	fmt.Println(dataJSON)
	newData, err := json.Marshal(dataJSON)
	if err != nil {
		return nil, err
	}
	return newData, nil
}

func updateTransaction(whereClause string, updParams []*proto_msg.UpdateCriteriaStruct, params []interface{}, tx *sql.Tx) error {
	var offset int64
	params = append(params, limit)
	params = append(params, offset)
	for {
		var count int64
		params[len(params)-1] = offset
		query := "SELECT id, data FROM gamelink.users WHERE " + whereClause + " LIMIT ? OFFSET ?"
		rows, err := tx.Query(query, params...)
		//defer rows.Close()
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}
		type update struct {
			id   int64
			data []byte
		}
		var updateSet []update
		for rows.Next() {
			count++
			var id int64
			var oldData []byte
			rows.Scan(&id, &oldData)
			newData, err := prepareDataToUpdateInDb(oldData, updParams)
			if err != nil {
				return err
			}
			upd := update{id: id, data: newData}
			updateSet = append(updateSet, upd)
		}
		for _, v := range updateSet {
			_, err = tx.Exec("UPDATE users set data = ? WHERE id = ?", v.data, v.id)
			if err != nil {
				return err
			}
		}
		if count < limit {
			break
		}
		offset += limit
	}
	return nil
}
