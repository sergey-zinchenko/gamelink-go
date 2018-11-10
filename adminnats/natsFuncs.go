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

//ConnectionNats - add nats connection to natsService struct
func (ns *NatsService) ConnectionNats(nc *nats.Conn) {
	ns.nc = nc
	//TODO: надо бы сюда перенести код подключения к натс
}

//PreparePushMessage - divides receivers into two arrays
func (ns *NatsService) PreparePushMessage(msg string, receivers []*push.UserInfo) error {
	var iosReceivers, androidReceivers []*push.UserInfo
	for _, v := range receivers {
		if v.DeviceOS == "ios" {
			iosReceivers = append(iosReceivers, v)
		} else if v.DeviceOS == "android" {
			androidReceivers = append(androidReceivers, v)
		}
	}
	if androidReceivers != nil {
		ns.sendAndroidPush(msg, androidReceivers)
	}
	//ns.sendIosPush(msg, iosReceivers)
	return nil
}

//TODO: Есть такое предположение, что нуно слить PreparePushMessage и sendAndroidPush и sendIosPush - отправлять пуши по одному и не париться - будет меньше кода

//sendAndroidPush - send push messages to android receivers
func (ns *NatsService) sendAndroidPush(msg string, receivers []*push.UserInfo) error {
	sendStruct := push.PushMsgStruct{Message: msg, UserInfo: receivers}
	data, err := proto.Marshal(&sendStruct)
	if err != nil {
		return err
	}
	if err := ns.nc.Publish(config.NatsAndroidChan, data); err != nil {
		return err
	}
	return nil
}

//func (ns *NatsService) sendIosPush(msg string, receivers []*push.UserInfo) error {
//	sendStruct := push.PushMsgStruct{Message: msg, UserInfo: receivers}
//	data, err := proto.Marshal(&sendStruct)
//	if err != nil {
//		return err
//	}
//	if err := ns.nc.Publish(config.NatsIosChan, data); err != nil {
//		return err
//	}
//	return nil
//}
