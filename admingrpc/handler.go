package admingrpc

import (
	"errors"
	"fmt"
	msg "gamelink-go/protoMsg"
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
		subquery string
		ctx      context.Context
		params   []*msg.OneCriteriaStruct
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
	return "", nil
}

func dateParser(v *msg.OneCriteriaStruct) (string, error) {
	q := "str_to_date(bdate, '%d.%m.%Y')"
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
	year := time.Now().Year() - y
	month := int(time.Now().Month())
	var val string
	if month < 10 {
		val = fmt.Sprintf("%d.0%d.%d", time.Now().Day(), month, year)
	} else {
		val = fmt.Sprintf("%d.%d.%d", time.Now().Day(), month, year)
	}

	q += "str_to_date(" + "\"" + val + "\"" + ", '%d.%m.%Y')"
	return q, nil
}
