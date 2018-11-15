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
		name, deviceID, msgSystem string
		receivers                 []*push.UserInfo
	)
	lb := fmt.Sprintf("lb%d", lbNum)
	logrus.Warn("rec lb: ", lb)
	logrus.Warn("rec score: ", newScore)
	logrus.Warn("id: ", u.ID())
	rows, err := u.dbs.mySQL.Query(queries.GetPushReceiversData, lb, u.ID(), lb, lb, u.ID(), lb, newScore)
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.Warn("no rows")
			return nil, nil
		}
		return nil, err
	}
	fmt.Println(err == sql.ErrNoRows)
	logrus.Warn("rows: ", rows)
	logrus.Warn("rowsERR: ", rows.Err())
	logrus.Warn("next ", rows.Next())
	for rows.Next() {
		logrus.Warn("next row")
		err = rows.Scan(&name, &deviceID, &msgSystem)
		if err != nil {
			return nil, err
		}
		var receiver *push.UserInfo
		if name != "" {
			receiver.Name = name
		}
		if deviceID != "" {
			receiver.DeviceID = deviceID
		}
		if msgSystem != "" {
			switch msgSystem {
			case push.UserInfo_apns.String():
				receiver.MsgSystem = push.UserInfo_apns
			case push.UserInfo_firebase.String():
				receiver.MsgSystem = push.UserInfo_firebase
			}
		}
		receivers = append(receivers, receiver)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		logrus.Warn(err.Error())
	}
	rows.Close()
	return receivers, nil
}
