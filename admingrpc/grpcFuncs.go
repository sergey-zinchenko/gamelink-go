package admingrpc

import (
	"fmt"
	msg "gamelink-go/protoMsg"
	"golang.org/x/net/context"
	"strconv"
)

type (
	//AdminServiceServer - grpc server struct
	AdminServiceServer struct{}
)

//Count - handle /count command from bot
func (s *AdminServiceServer) Count(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.CountResponse, error) {
	var count int
	subquery := "SELECT COUNT(id) FROM users WHERE "
	h := Handler{subquery, ctx, in.GetParams()}
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
	fmt.Println(subquery + query)
	data = "123123123123123123123123123"
	count, err = strconv.Atoi(data)
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
	//Реализация метода
	return &msg.OneUserResponse{User: user}, nil
}

//func parser(in *protoMsg.MultiCriteriaRequest) (string, error) {
//	var subQuery string
//	for k, v := range in.GetParams() {
//		if k > 0 {
//			subQuery += " AND "
//		}
//		if v.Cr == protoMsg.OneCriteriaStruct_age {
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
//		case protoMsg.OneCriteriaStruct_l:
//			subQuery += " < "
//		case protoMsg.OneCriteriaStruct_e:
//			subQuery += " = "
//		case protoMsg.OneCriteriaStruct_g:
//			subQuery += " > "
//		}
//		subQuery += "\"" + v.Value + "\""
//	}
//	return subQuery, nil
//}
//
//func dateParser(v *protoMsg.OneCriteriaStruct) (string, error) {
//	q := "str_to_date(bdate, '%d.%m.%Y')"
//	switch v.Op {
//	case protoMsg.OneCriteriaStruct_l:
//		q += " > "
//	case protoMsg.OneCriteriaStruct_e:
//		q += " = "
//	case protoMsg.OneCriteriaStruct_g:
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
