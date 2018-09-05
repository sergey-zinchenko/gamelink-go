package admin

import (
	"gamelink-go/prot"
	"golang.org/x/net/context"
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
		params   []*prot.OneCriteriaStruct
	}
)

//CheckCtx - check ctx timeout
func (h Handler) CheckCtx() error {
	return nil
}

//ParseParams - parse params from request and return db query string
func (h Handler) ParseParams() (string, error) {
	return "", nil

}

//GetData - get data from db
func (h Handler) GetData(query string) (string, error) {
	return "", nil
}
