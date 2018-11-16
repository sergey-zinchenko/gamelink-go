package storage

import (
	"database/sql"
	"fmt"
	push "gamelink-go/proto_nats_msg"
	"gamelink-go/storage/queries"
	"github.com/sirupsen/logrus"
)

//GetPushReceivers - returns array of receivers, which the user is ahead on points
func (u User) GetPushReceivers(newScore int, lbNum int) ([]*push.UserInfo, error) {
	var (
		name, deviceID, msgSystem sql.NullString
		receivers                 []*push.UserInfo
	)
	lb := fmt.Sprintf("lb%d", lbNum)
	qs := fmt.Sprintf(queries.GetPushReceiversData, lb, lb, lb, lb)
	rows, err := u.dbs.mySQL.Query(qs, u.ID(), u.ID(), newScore)
	defer rows.Close()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	for rows.Next() {
		var receiver push.UserInfo
		err = rows.Scan(&name, &deviceID, &msgSystem)
		if err != nil {
			logrus.Warn(err.Error())
			return nil, err
		}
		if name.Valid {
			receiver.Name = name.String
		}
		if deviceID.Valid {
			receiver.DeviceID = deviceID.String
		}
		if msgSystem.Valid {
			switch msgSystem.String {
			case push.UserInfo_apns.String():
				receiver.MsgSystem = push.UserInfo_apns
			case push.UserInfo_firebase.String():
				receiver.MsgSystem = push.UserInfo_firebase
			}
		}
		receivers = append(receivers, &receiver)
	}
	return receivers, nil
}
