package admingrpc

import (
	"encoding/json"
	"gamelink-go/proto_msg"
	"time"
)

func convertToGrpcStruct(res []interface{}) ([]*proto_msg.UserResponseStruct, error) {
	var users []*proto_msg.UserResponseStruct
	for _, v := range res {
		var u map[string]interface{}
		r := []byte(v.(string))
		err := json.Unmarshal(r, &u)
		if err != nil {
			return nil, err
		}
		user := mapToStruct(u)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func mapToStruct(u map[string]interface{}) *proto_msg.UserResponseStruct {
	var user proto_msg.UserResponseStruct
	if u["id"] != nil {
		user.Id = int64(u["id"].(float64))
	}
	if u["vk_id"] != nil {
		user.VkId = u["vk_id"].(string)
	}
	if u["fb_id"] != nil {
		user.FbId = u["fb_id"].(string)
	}
	if u["name"] != nil {
		user.Name = u["name"].(string)
	}
	if u["country"] != nil {
		user.Country = u["country"].(string)
	}
	if u["sex"] != nil {
		if u["sex"] == "M" {
			user.Sex = proto_msg.UserResponseStruct_M
		} else if u["sex"] == "M" {
			user.Sex = proto_msg.UserResponseStruct_F
		}
	}
	if u["bdate"] != nil {
		user.Age = int64(u["bdate"].(float64))
	}
	if u["created_at"] != nil {
		user.CreatedAt = time.Unix(int64(u["created_at"].(float64)), 0).Format(time.ANSIC)
	}
	if u["deleted"] != nil {
		user.Deleted = int32(u["deleted"].(float64))
	}
	if u["email"] != nil {
		user.Email = u["email"].(string)
	}
	return &user
}
