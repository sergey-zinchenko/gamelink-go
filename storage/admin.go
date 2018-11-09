package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/proto_msg"
	"github.com/kataras/iris/core/errors"
)

const (
	//QueryCount - const for count query
	QueryCount QueryKind = iota
	//QuerySelect - const for Select query
	QuerySelect
	//QuerySelectWithDeviceJoin - const for select query with join
	QuerySelectWithDeviceJoin
	//QueryUpdate - const for update query
	QueryUpdate
	//QueryDelete - const for delete query
	QueryDelete

	limit = 100
)

type (
	//QueryBuilder - struct for work with params and build db query
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
	//update - struct contains userID and data, prepared to update in db
	update struct {
		id   int64
		data []byte
	}
)

//WithClause - func for make query part from criteria
func (q *QueryBuilder) WithClause(criteria *proto_msg.OneCriteriaStruct) {
	if q.whereClause != "" {
		q.whereClause += " AND "
	}
	if criteria.Cr == proto_msg.OneCriteriaStruct_updated_at || criteria.Cr == proto_msg.OneCriteriaStruct_created_at {
		q.whereClause += "unix_timestamp(" + criteria.Cr.String() + ") "
	} else if criteria.Cr == proto_msg.OneCriteriaStruct_age {
		q.whereClause += " timestampdiff(YEAR, from_unixtime(bdate), curdate()) "
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

//SelectQueryWithDeviceJoin - first part of select query with join to deviceID table
func (q QueryBuilder) SelectQueryWithDeviceJoin() QueryBuilder {
	q.kind = QuerySelectWithDeviceJoin
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
	case QuerySelect, QueryCount, QueryDelete, QuerySelectWithDeviceJoin:
		var query string
		switch q.kind {
		case QueryCount:
			query = `SELECT COUNT(id) FROM users`
		case QuerySelect:
			query = `SELECT id, vk_id, fb_id, name, email, sex, timestampdiff(YEAR, bdate, curdate()), country, date(created_at), deleted from users`
		case QuerySelectWithDeviceJoin:
			query = `SELECT name, device_id, device_os from users LEFT JOIN device_ids ON id=user_id`
		case QueryDelete:
			query = `UPDATE users SET deleted=1`
		}
		if q.whereClause != "" {
			query = fmt.Sprintf("%s WHERE %s", query, q.whereClause)
		}
		rows, err := mysql.Query(query, q.params...)
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
		rows.Close()
		if res == nil {
			return nil, errors.New("there is no users satisfied input params")
		}
		return res, nil
	case QueryUpdate:
		tx, err := mysql.Begin()
		if err != nil {
			return nil, err
		}
		err = q.updateTransaction(tx)
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
func prepareDataToUpdateInDb(id int64, rowData []byte, updateParams []*proto_msg.UpdateCriteriaStruct) (*update, error) {
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
	newData, err := json.Marshal(dataJSON)
	if err != nil {
		return nil, err
	}
	return &update{id: id, data: newData}, nil
}

func (q QueryBuilder) updateTransaction(tx *sql.Tx) error {
	var offset int64
	q.params = append(q.params, limit)
	q.params = append(q.params, 0) //append offset == 0
	for {
		var count int64
		q.params[len(q.params)-1] = offset
		query := "SELECT id, data FROM gamelink.users WHERE " + q.whereClause + " LIMIT ? OFFSET ?"
		rows, err := tx.Query(query, q.params...)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}

		var updateSet []*update
		for rows.Next() {
			count++
			var id int64
			var oldData []byte
			err = rows.Scan(&id, &oldData)
			if err != nil {
				return err
			}
			updated, err := prepareDataToUpdateInDb(id, oldData, q.updParams)
			if err != nil {
				return err
			}
			updateSet = append(updateSet, updated)
		}
		rows.Close()
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
