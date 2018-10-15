package admingrpc

import (
	"database/sql"
	"errors"
	msg "gamelink-go/proto_msg"
	push "gamelink-go/proto_nats_msg"
	"gamelink-go/storage"
	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
	"golang.org/x/net/context"
)

type (
	//AdminServiceServer - grpc server struct
	AdminServiceServer struct {
		dbs *storage.DBS
		nc  *nats.Conn
	}
)

const (
	androidNatsChannel = "android_push"
	iosNatsChannel     = "ios_push"
)

//Dbs - set dbs to adminServiceServer
func (s *AdminServiceServer) Dbs(dbs *storage.DBS) {
	s.dbs = dbs
}

//Nats - set nats connection to adminServiceServer
func (s *AdminServiceServer) Nats(nc *nats.Conn) {
	s.nc = nc
}

//Count - handle /count command from bot
func (s *AdminServiceServer) Count(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.CountResponse, error) {
	b := storage.QueryBuilder{}.CountQuery().WithMultipleClause(in.Params)
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
	var users []*msg.UserResponseStruct
	b := storage.QueryBuilder{}.SelectQuery().WithMultipleClause(in.Params)
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
	return &msg.MultiUserResponse{Users: users}, nil
}

//Update - handle /update command from bot
func (s *AdminServiceServer) Update(ctx context.Context, in *msg.UpdateCriteriaRequest) (*msg.StringResponse, error) {
	b := storage.QueryBuilder{}.UpdateQuery().WithMultipleClause(in.FindParams).WithData(in.UpdParams)
	_, err := s.dbs.Query(b, func(scanFunc storage.ScanFunc) (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return &msg.StringResponse{Response: "success"}, nil
}

//Delete - handle /delete command from bot
func (s *AdminServiceServer) Delete(ctx context.Context, in *msg.MultiCriteriaRequest) (*msg.OneUserResponse, error) {
	b := storage.QueryBuilder{}.DeleteQuery().WithMultipleClause(in.Params)
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
func (s *AdminServiceServer) SendPush(ctx context.Context, in *msg.PushCriteriaRequest) (*msg.StringResponse, error) {
	var ios, android []*push.UserInfo
	b := storage.QueryBuilder{}.SelectQueryWithDeviceJoin().WithMultipleClause(in.Params)
	_, err := s.dbs.Query(b, func(scanFunc storage.ScanFunc) (interface{}, error) {
		var name, deviceID, deviceOs sql.NullString
		err := scanFunc(&name, &deviceID, &deviceOs)
		if err != nil {
			return nil, err
		}
		var info push.UserInfo
		if name.Valid {
			info.Name = name.String
		}
		if deviceID.Valid {
			info.DeviceID = deviceID.String
		}
		if deviceOs.Valid {
			switch deviceOs.String {
			case "android":
				android = append(android, &info)
			case "ios":
				ios = append(ios, &info)
			}
		}
		return info, nil
	})
	if err != nil {
		return nil, err
	}
	if ios != nil {
		err = s.sendPushByNats(iosNatsChannel, in.Message, ios)
	}
	if err != nil {
		return nil, err
	}
	if android != nil {
		err = s.sendPushByNats(androidNatsChannel, in.Message, android)
	}
	if err != nil {
		return nil, err
	}
	return &msg.StringResponse{Response: "message successfully send"}, nil
}

func (s *AdminServiceServer) sendPushByNats(subject string, msg string, receivers []*push.UserInfo) error {
	sendStruct := push.PushMsgStruct{Message: msg, UserInfo: receivers}
	data, err := proto.Marshal(&sendStruct)
	if err != nil {
		return err
	}
	if err := s.nc.Publish(subject, data); err != nil {
		return err
	}
	return nil
}
