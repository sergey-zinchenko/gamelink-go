package adminnats

import (
	"gamelink-go/config"
	push "gamelink-go/proto_nats_msg"
	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
)

//NatsService - struct for nats connection
type NatsService struct {
	nc *nats.Conn
}

//ConnectNats - add nats connection to natsService struct
func ConnectNats() (*NatsService, error) {
	connats := NatsService{}
	nc, err := nats.Connect(config.NATSPort)
	if err != nil {
		return nil, err
	}
	connats.nc = nc
	return &connats, nil
}

//PreparePushMessage - divides receivers into two arrays
func (ns *NatsService) PreparePushMessage(msg string, receivers []*push.UserInfo) error {
	for _, v := range receivers {
		sendStruct := push.PushMsgStruct{Message: msg, UserInfo: v}
		data, err := proto.Marshal(&sendStruct)
		if err != nil {
			return err
		}
		if v.DeviceOS == "ios" {
			if err := ns.nc.Publish(config.NatsIosChan, data); err != nil {
				return err
			}
		} else if v.DeviceOS == "android" {
			if err := ns.nc.Publish(config.NatsAndroidChan, data); err != nil {
				return err
			}
		}
	}
	return nil
}
