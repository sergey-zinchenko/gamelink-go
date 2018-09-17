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
		user.Age = (time.Now().Unix() - int64(u["bdate"].(float64))) / 31556926 // Но это неверно т.к. не учтены високосные года....надо допилить
	}
	if u["created_at"] != nil {
		user.CreatedAt = time.Unix(u["created_at"].(int64), 0).Format(time.RFC3339)
	}
	if u["deleted"] != nil {
		user.Deleted = u["deleted"].(int32)
	}
	if u["email"] != nil {
		user.Email = u["email"].(string)
	}
	return &user
}
