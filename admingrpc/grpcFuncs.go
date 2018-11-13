package admingrpc

import (
	"database/sql"
	"errors"
	"gamelink-go/adminnats"
	"gamelink-go/config"
	msg "gamelink-go/proto_msg"
	push "gamelink-go/proto_nats_msg"
	service "gamelink-go/proto_service"
	"gamelink-go/storage"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type (
	//AdminServiceServer - grpc server struct
	AdminServiceServer struct {
		dbs *storage.DBS
		nc  *adminnats.NatsService
	}
)

//Connect - set grpc connection to adminServiceServer
func (s *AdminServiceServer) Connect() error {
	lis, err := net.Listen(config.GRPCNetwork, config.GRPCPort)
	if err != nil {
		return err
	}
	grpcServ := grpc.NewServer()
	service.RegisterAdminServiceServer(grpcServ, s)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServ)
	if err := grpcServ.Serve(lis); err != nil {
		return err
	}
	return nil
}

//SetDbsToAdminService - set dbs to adminServiceServer
func (s *AdminServiceServer) SetDbsToAdminService(dbs *storage.DBS) {
	s.dbs = dbs
}

//SetNatsToAdminService - set nats connection to adminServiceServer
func (s *AdminServiceServer) SetNatsToAdminService(nc *adminnats.NatsService) {
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
	var users []*push.UserInfo
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
			info.DeviceOS = deviceOs.String
		}
		if info.DeviceID != "" && info.DeviceOS != "" && info.Name != "" {
			users = append(users, &info)
		}
		return info, nil
	})
	if err != nil {
		return nil, err
	}
	err = s.nc.PreparePushMessage(in.Message, users)
	if err != nil {
		return nil, err
	}
	return &msg.StringResponse{Response: "message successfully send"}, nil
}
