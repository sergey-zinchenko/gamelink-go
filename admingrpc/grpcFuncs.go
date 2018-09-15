package admingrpc

import (
	"fmt"
	msg "gamelink-go/proto_msg"
	"gamelink-go/storage"
	"golang.org/x/net/context"
	"strconv"
)

type (
	//AdminServiceServer - grpc server struct
	AdminServiceServer struct {
		dbs *storage.DBS
	}
)

//Dbs - set dbs to adminServiceServer
func (s *AdminServiceServer) Dbs(dbs *storage.DBS) {
	s.dbs = dbs
}

//Count - handle /count command from bot
func (s *AdminServiceServer) Count(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.CountResponse, error) {
	var count int
	h := Handler{s.dbs, ctx, in.GetParams(), "count"}
	err := h.CheckCtx()
	if err != nil {
		return nil, err
	}
	query, err := h.ParseParams()
	if err != nil {
		return nil, err
	}
	data, err := h.GetData(query)
	if err != nil {
		return nil, err
	}
	count, err = strconv.Atoi(data)
	fmt.Println(count)
	fmt.Println(int64(0))
	//TODO: разобраться с ответом 0 т.к. в структуре ответа стоит omitempty и 0 не возвращается. Но это не точно!!!
	return &msg.CountResponse{Count: int64(count)}, nil
}

//Find - handle /find command from bot
func (s *AdminServiceServer) Find(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.MultiUserResponse, error) {
	var users []*msg.UserResponseStruct
	return &msg.MultiUserResponse{Users: users}, nil
}

//Update - handle /update command from bot
func (s *AdminServiceServer) Update(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.MultiUserResponse, error) {
	var users []*msg.UserResponseStruct
	//Реализация метода
	return &msg.MultiUserResponse{Users: users}, nil
}

//Delete - hande /delete command from bot
func (s *AdminServiceServer) Delete(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.OneUserResponse, error) {
	var user *msg.UserResponseStruct
	h := Handler{s.dbs, ctx, in.GetParams(), "delete"}
	err := h.CheckCtx()
	if err != nil {
		return nil, err
	}
	query, err := h.ParseParams()
	if err != nil {
		return nil, err
	}
	data, err := h.GetData(query)
	if err != nil {
		return nil, err
	}
	fmt.Println(data)
	//Тут нужно собрать из ответа из бд структуру ответа (OneUserResponse)
	return &msg.OneUserResponse{User: user}, nil
}

//func parser(in *proto_msg.MultiCriteriaRequest) (string, error) {
//	var subQuery string
//	for k, v := range in.GetParams() {
//		if k > 0 {
//			subQuery += " AND "
//		}
//		if v.Cr == proto_msg.OneCriteriaStruct_age {
//			q, err := dateParser(v)
//			if err != nil {
//				return "", err
//			}
//			subQuery += q
//			continue
//		} else {
//			subQuery += v.Cr.String()
//		}
//		switch v.Op {
//		case proto_msg.OneCriteriaStruct_l:
//			subQuery += " < "
//		case proto_msg.OneCriteriaStruct_e:
//			subQuery += " = "
//		case proto_msg.OneCriteriaStruct_g:
//			subQuery += " > "
//		}
//		subQuery += "\"" + v.Value + "\""
//	}
//	return subQuery, nil
//}
//
//func dateParser(v *proto_msg.OneCriteriaStruct) (string, error) {
//	q := "str_to_date(bdate, '%d.%m.%Y')"
//	switch v.Op {
//	case proto_msg.OneCriteriaStruct_l:
//		q += " > "
//	case proto_msg.OneCriteriaStruct_e:
//		q += " = "
//	case proto_msg.OneCriteriaStruct_g:
//		q += " < "
//	}
//	y, err := strconv.Atoi(v.Value)
//	if err != nil {
//		return "", err
//	}
//	year := time.Now().Year() - y
//	month := int(time.Now().Month())
//	var val string
//	if month < 10 {
//		val = fmt.Sprintf("%d.0%d.%d", time.Now().Day(), month, year)
//	} else {
//		val = fmt.Sprintf("%d.%d.%d", time.Now().Day(), month, year)
//	}
//
//	q += "str_to_date(" + "\"" + val + "\"" + ", '%d.%m.%Y')"
//	return q, nil
//}
