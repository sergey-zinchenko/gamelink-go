package admingrpc

import (
	"database/sql"
	"errors"
	"fmt"
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
	var users []*msg.UserResponseStruct
	b.SelectQuery().WithMultipleClause(in.Params)
	_, err := s.dbs.Query(b, func(scanFunc storage.ScanFunc) (interface{}, error) {
		var (
			id, age                                          sql.NullInt64
			vkID, fbID, name, email, sex, country, createdAt sql.NullString
			deleted                                          sql.NullInt64
		)
		err := scanFunc(&id, &vkID, &fbID, &name, &email, &sex, &age, &country, &createdAt, &deleted)
		if err != nil {
			return nil, err
		}
		var user msg.UserResponseStruct
		if id.Valid {
			user.Id = id.Int64
		}
		if vkID.Valid {
			user.VkId = vkID.String
		}
		if fbID.Valid {
			user.FbId = fbID.String
		}
		if name.Valid {
			user.Name = name.String
		}
		if country.Valid {
			user.Country = country.String
		}
		if sex.Valid {
			if sex.String == "M" {
				user.Sex = msg.UserResponseStruct_M
			} else {
				user.Sex = msg.UserResponseStruct_F
			}
		}
		if age.Valid {
			user.Age = age.Int64
		}
		if createdAt.Valid {
			user.CreatedAt = createdAt.String
		}
		if deleted.Valid {
			user.Deleted = int32(deleted.Int64)
		}
		if email.Valid {
			user.Email = email.String
		}
		users = append(users, &user)
		return user, nil
	})
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &msg.MultiUserResponse{Users: users}, nil
}

//Update - handle /update command from bot
func (s *AdminServiceServer) Update(ctx context.Context, in *msg.UpdateCriteriaRequest) (*msg.StringResponse, error) {
	count, err := s.Count(ctx, &msg.MultiCriteriaRequest{Params: in.FindParams})
	if err != nil {
		return nil, err
	}
	if count.Count == 0 {
		return nil, errors.New("there is no users satisfy input params")
	}
	var i int64
	for i = 0; i < count.Count; i = i + 1 {
		g := storage.QueryBuilder{}
		g.Offset(i)
		g.GetData().WithMultipleClause(in.FindParams)
		_, err = s.dbs.Query(g, func(scanFunc storage.ScanFunc) (interface{}, error) {
			var bytes []byte
			var ident int64
			err := scanFunc(&ident, &bytes)
			if err != nil {
				return nil, err
			}
			u := storage.UpdateBuilder{ID: ident, Data: bytes, UpdParams: in.UpdParams}
			updated, err := u.Prepare()
			if err != nil {
				return nil, err
			}
			err = s.dbs.Update(updated)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	}
	if err != nil {
		return nil, err
	}
	return &msg.StringResponse{Response: "success"}, nil // Стоит ли тут возвращать массив юзеров? И если да, доделать этот момент
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
	fmt.Println(in.Params)
	b := storage.QueryBuilder{}
	b.PushQuery().WithMultipleClause(in.Params)
	//обрабытваем то шо нашли по запросу из базы
	fmt.Println(b.Message())
	return &msg.StringResponse{Response: "message successfully send"}, nil
}
