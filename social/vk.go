package social


//TODO: нужно обработать {"error":"d=vk; c=[28]; m=Application authorization failed: refresh service token"}

import (
	"net/http"
	"gamelink-go/config"
	"encoding/json"
	"gamelink-go/graceful"
	"sync"
	log "github.com/sirupsen/logrus"
	"fmt"
)

type (
	VkToken struct {
		token string
	}

	vkError struct {
		Code int `json:"error_code"`
		Message string `json:"error_msg"`
	}
)

var serviceKey string
var once sync.Once

func NewVkToken(token string) *VkToken {
	return &VkToken{token}
}

func requestServiceKey() string {
	log.Debug("vk.requestServiceKey")
	var requestFunc = func() (string, *graceful.Error) {
		log.Debug("vk.requestServiceKey.requestFunc")
		type (

			vkAccessTokenResponse struct {
				AccessToken string `json:"access_token"`
				Error string `json:"error"`
				ErrorDescription string `json:"error_description"`
			}
		)
		req, err := http.NewRequest("GET", "https://oauth.vk.com/access_token", nil)
		if err != nil {
			return "", graceful.NewNetworkError(err.Error())
		}
		q := req.URL.Query()
		q.Add("client_id", config.VkontakteAppId)
		q.Add("client_secret", config.VkontakteAppSecret)
		q.Add("v", "5.68")
		q.Add("grant_type", "client_credentials")
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			return "", graceful.NewNetworkError(err.Error())
		}
		defer resp.Body.Close()
		var f vkAccessTokenResponse
		err = json.NewDecoder(resp.Body).Decode(&f)
		if err != nil {
			return "", graceful.NewParsingError(err.Error())
		}
		if f.Error != "" {
			return "", graceful.NewVkError(f.ErrorDescription)
		}
		if f.AccessToken == "" {
			return "", graceful.NewInvalidError("empty access_token")
		}
		return f.AccessToken, nil
	}
	once.Do(func() {
		var err *graceful.Error = nil
		if serviceKey, err = requestFunc(); err != nil {
			log.WithError(err).Fatal("cant get vk service key")
		}
	})
	return serviceKey
}

func (vk VkToken) checkToken() (string, *graceful.Error) {
	type (
		vkCheckTokenData struct {
			Success int `json:"success"`
			UserId int64 `json:"user_id"`
		}

		vkCheckTokenResponse struct {
			Response vkCheckTokenData `json:"response"`
			Error *vkError `json:"error"`
		}
	)
	log.Debug("vk.checkToken")
	req, err := http.NewRequest("GET", "https://api.vk.com/method/secure.checkToken", nil)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("access_token", requestServiceKey())
	q.Add("client_secret", config.VkontakteAppSecret)
	q.Add("token", vk.token)
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	defer resp.Body.Close()
	var f vkCheckTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", graceful.NewParsingError(err.Error())
	}
	if f.Error != nil {
		switch f.Error.Code {
		case 15:
			return "", graceful.NewNotFoundError(f.Error.Message, f.Error.Code)
		default:
			return "", graceful.NewVkError(f.Error.Message, f.Error.Code)
		}
	}
	if f.Response.Success != 1 {
		return "", graceful.NewNotFoundError("bad success flag")
	}
	if f.Response.UserId == 0 {
		return "", graceful.NewInvalidError("empty user id")
	}
	return fmt.Sprint(f.Response.UserId), nil
}

func (vk VkToken) get(userId string) (string, *graceful.Error) {
	type (
		vkUsersGetData struct {
			FirstName string `json:"first_name"`
			LastName string `json:"last_name"`
			Id int64 `json:"id"`
		}

		vkUsersGetResponse struct {
			Response []vkUsersGetData `json:"response"`
			Error *vkError `json:"error"`
		}
	)
	log.Debug("vk.get")
	req, err := http.NewRequest("GET", "https://api.vk.com/method/users.get", nil)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("fields", "sex,bdate,city,country")
	q.Add("user_ids", userId)
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	defer resp.Body.Close()
	var f vkUsersGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", graceful.NewParsingError(err.Error())
	}
	if f.Error != nil {
		return "", graceful.NewVkError(f.Error.Message, f.Error.Code)
	}
	if len(f.Response) != 1 || fmt.Sprint(f.Response[0].Id) != userId  {
		return "", graceful.NewInvalidError("user id not match or empty response")
	}
	return f.Response[0].FirstName + " " + f.Response[0].LastName, nil
}

func (vk VkToken) GetUserInfo() (string, string, *graceful.Error) {
	log.Debug("vk.GetUserInfo")
	id, err := vk.checkToken()
	if err != nil {
		log.WithError(err).Debug("vk token failed")
		return "", "", err
	}
	log.WithField("vk_id", id).Debug("vk token ok")
	name, err := vk.get(id)
	if err != nil {
		log.WithError(err).Debug("vk user failed")
		return id, "", err
	}
	log.WithField("name", name).Debug("vk user ok")
	return id, name, nil
}