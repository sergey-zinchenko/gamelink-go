package admingrpc

import (
	"errors"
	msg "gamelink-go/proto_msg"
	"gamelink-go/storage"
	"golang.org/x/net/context"
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
	b := storage.QueryBuilder{}
	b.CountQuery().WithMultipleClause(in.Params)
	res, err := s.dbs.Query(b, func(scanFunc storage.ScanFunc) (interface{}, error) {
		var countresp int64
		err := scanFunc(&countresp)
		if err != nil {
			return nil, err
		}
		return countresp, nil
	})
	if err != nil {
		return nil, err
	}
	r, ok := res[0].(int64)
	if !ok {
		return nil, errors.New("can't convert to int")
	}
	return &msg.CountResponse{Count: r}, nil
}

//Find - handle /find command from bot
func (s *AdminServiceServer) Find(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.MultiUserResponse, error) {
	b := storage.QueryBuilder{}
	b.SelectQuery().WithMultipleClause(in.Params)
	res, err := s.dbs.Query(b, func(scanFunc storage.ScanFunc) (interface{}, error) {
		var findres string
		err := scanFunc(&findres)
		if err != nil {
			return nil, err
		}
		return findres, nil
	})
	if err != nil {
		return nil, err
	}
	users, err := convertToGrpcStruct(res)
	if err != nil {
		return nil, err
	}
	return &msg.MultiUserResponse{Users: users}, nil
}

//Update - handle /update command from bot
func (s *AdminServiceServer) Update(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.MultiUserResponse, error) {
	var users []*msg.UserResponseStruct
	//Реализация метода
	return &msg.MultiUserResponse{Users: users}, nil
}

//Delete - handle /delete command from bot
func (s *AdminServiceServer) Delete(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.OneUserResponse, error) {
	b := storage.QueryBuilder{}
	b.DeleteQuery().WithMultipleClause(in.Params)
	_, err := s.dbs.Query(b, func(scanFunc storage.ScanFunc) (interface{}, error) {
		var deleted interface{}
		err := scanFunc(&deleted)
		if err != nil {
			return nil, err
		}
		return deleted, nil
	})
	if err != nil {
		return nil, err
	}
	users, err := s.Find(ctx, in)
	if err != nil {
		return nil, err
	}
	if users.Users[0] == nil {
		return nil, errors.New("user not found")
	}
	return &msg.OneUserResponse{User: users.Users[0]}, nil
}

//SendPush - handle /send_push command
func (s *AdminServiceServer) SendPush(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.StringResponse, error) {
	//
	//
	//
	return &msg.StringResponse{Response: "message successfully send"}, nil
}
