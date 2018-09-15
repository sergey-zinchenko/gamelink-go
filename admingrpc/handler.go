package admingrpc

import (
	"errors"
	msg "gamelink-go/protoMsg"
	"gamelink-go/storage"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

type (
	//Handle - handler interface
	Handle interface {
		CheckCtx() error
		ParseParams() (string, error)
		GetData(query string) (string, error)
	}
	//Handler - handle interface struct
	Handler struct {
		dbs     *storage.DBS
		ctx     context.Context
		params  []*msg.OneCriteriaStruct
		command string
	}
)

//CheckCtx - check ctx timeout
func (h Handler) CheckCtx() error {
	if h.ctx.Err() == context.Canceled {
		return errors.New("client cancelled, abandoning")
	}
	return nil
}

//ParseParams - parse params from request and return db query string
func (h Handler) ParseParams() (string, error) {
	var subQuery string
	for k, v := range h.params {
		if k > 0 {
			subQuery += " AND "
		}
		if v.Cr == msg.OneCriteriaStruct_age {
			q, err := dateParser(v)
			if err != nil {
				return "", err
			}
			subQuery += q
			continue
		} else {
			subQuery += v.Cr.String()
		}
		switch v.Op {
		case msg.OneCriteriaStruct_l:
			subQuery += " < "
		case msg.OneCriteriaStruct_e:
			subQuery += " = "
		case msg.OneCriteriaStruct_g:
			subQuery += " > "
		}
		subQuery += "\"" + v.Value + "\""
	}
	return subQuery, nil

}

//GetData - get data from db
func (h Handler) GetData(query string) (string, error) {
	var res string
	var err error
	a := h.dbs.Admin()
	switch h.command {
	case "count":
		res, err = a.Count(query)
	case "delete":
		res, err = a.Delete(query)
	}
	if err != nil {
		return "", err
	}
	return res, nil
}

//dateParser - parse date params
func dateParser(v *msg.OneCriteriaStruct) (string, error) {
	q := "unix_timestamp(users.bdate)"
	switch v.Op {
	case msg.OneCriteriaStruct_l:
		q += " > "
	case msg.OneCriteriaStruct_e:
		q += " = "
	case msg.OneCriteriaStruct_g:
		q += " < "
	}
	y, err := strconv.Atoi(v.Value)
	if err != nil {
		return "", err
	}
	t := time.Now()
	t = t.AddDate(-y, 0, 0)
	var val string
	val = strconv.Itoa(int(t.Unix()))
	q += val
	return q, nil
}
