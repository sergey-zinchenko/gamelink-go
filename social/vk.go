package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"gamelink-go/config"
	"gamelink-go/graceful"
	"net/http"
	"sync"
)

type (
	vkError struct {
		Code    int    `json:"error_code"`
		Message string `json:"error_msg"`
	}

	//VkToken - Class to get information and check validity about Vkontakte user tokens
	VkToken struct {
		token string
	}

	//VkServiceKey - structure for store and request vk service key
	VkServiceKey struct {
		key string
		m   sync.Mutex
	}
)

var serviceKey VkServiceKey

const (
	maxRetries = 10
)

//Key - method returns stored service key and request it from server if needed
func (sk *VkServiceKey) Key() (string, error) {
	sk.m.Lock()
	defer sk.m.Unlock()
	if sk.key != "" {
		return sk.key, nil
	}
	type (
		vkAccessTokenResponse struct {
			AccessToken      string `json:"access_token"`
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
	)
	req, err := http.NewRequest("GET", "https://oauth.vk.com/access_token", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("client_id", config.VkontakteAppID)
	q.Add("client_secret", config.VkontakteAppSecret)
	q.Add("v", "5.68")
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var f vkAccessTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", err
	}
	if f.Error != "" {
		return "", graceful.NewVkError(f.ErrorDescription)
	}
	if f.AccessToken == "" {
		return "", errors.New("empty access_token")
	}
	sk.key = f.AccessToken
	return sk.key, nil
}

//Reset - method removes stored service key and new one will be requested from server on next Key() call
func (sk *VkServiceKey) Reset() {
	sk.m.Lock()
	defer sk.m.Unlock()
	sk.key = ""
}

//NewVkToken - VkToken constructor
func NewVkToken(token string) *VkToken {
	return &VkToken{token}
}

func (vk VkToken) checkToken() (userID string, err error) {
	type (
		vkCheckTokenData struct {
			Success int   `json:"success"`
			UserID  int64 `json:"user_id"`
		}

		vkCheckTokenResponse struct {
			Response vkCheckTokenData `json:"response"`
			Error    *vkError         `json:"error"`
		}
	)
	var f vkCheckTokenResponse
	for i := 0; i < maxRetries; i++ {
		var accessToken string
		accessToken, err = serviceKey.Key()
		if err != nil {
			break
		}
		var req *http.Request
		req, err = http.NewRequest("GET", "https://api.vk.com/method/secure.checkToken", nil)
		if err != nil {
			break
		}
		q := req.URL.Query()
		q.Add("access_token", accessToken)
		q.Add("client_secret", config.VkontakteAppSecret)
		q.Add("token", vk.token)
		q.Add("v", "5.68")
		req.URL.RawQuery = q.Encode()
		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			break
		}
		err = json.NewDecoder(resp.Body).Decode(&f)
		resp.Body.Close() //defer replacement due to resource leak
		if err != nil {
			break
		}
		if f.Error != nil {
			switch f.Error.Code {
			case 15:
				err = &graceful.GracefulUnauthorizedError{}
			case 28: //обработка {"error":"d=vk; c=[28]; m=Application authorization failed: refresh service token"}
				serviceKey.Reset()
				fallthrough
			default:
				err = graceful.NewVkError(f.Error.Message, f.Error.Code)
			}
		}
	}
	if err != nil {
		return
	}
	if f.Response.Success != 1 {
		err = &graceful.GracefulUnauthorizedError{}
		return
	}
	if f.Response.UserID == 0 {
		err = errors.New("empty user id")
		return
	}
	userID = fmt.Sprint(f.Response.UserID)
	return
}

func (vk VkToken) get(userID string) (string, error) {
	type (
		vkUsersGetData struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			ID        int64  `json:"id"`
		}

		vkUsersGetResponse struct {
			Response []vkUsersGetData `json:"response"`
			Error    *vkError         `json:"error"`
		}
	)
	req, err := http.NewRequest("GET", "https://api.vk.com/method/users.get", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("fields", "sex,bdate,city,country")
	q.Add("user_ids", userID)
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var f vkUsersGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", err
	}
	if f.Error != nil {
		return "", graceful.NewVkError(f.Error.Message, f.Error.Code)
	}
	if len(f.Response) != 1 || fmt.Sprint(f.Response[0].ID) != userID {
		return "", errors.New("user id not match or empty response")
	}
	return f.Response[0].FirstName + " " + f.Response[0].LastName, nil
}

//GetUserInfo - method to check validity and get user information about the token if it valid. Returns NotFound error if token is not valid
func (vk VkToken) GetUserInfo() (string, string, error) {
	id, err := vk.checkToken()
	if err != nil {
		return "", "", err
	}
	name, err := vk.get(id)
	if err != nil {
		return id, "", err
	}
	return id, name, nil
}
