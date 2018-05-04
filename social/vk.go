package social

import (
	"encoding/json"
	"fmt"
	"gamelink-go/config"
	"gamelink-go/graceful"
	log "github.com/sirupsen/logrus"
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
func (sk *VkServiceKey) Key() (string, *graceful.Error) {
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
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("client_id", config.VkontakteAppID)
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

func (vk VkToken) checkToken() (userID string, ge *graceful.Error) {
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
	log.Debug("vk.checkToken")
	var f vkCheckTokenResponse
	for i := 0; i < maxRetries; i++ {
		var accessToken string
		accessToken, ge = serviceKey.Key()
		if ge != nil {
			break
		}
		req, err := http.NewRequest("GET", "https://api.vk.com/method/secure.checkToken", nil)
		if err != nil {
			ge = graceful.NewNetworkError(err.Error())
			break
		}
		q := req.URL.Query()
		q.Add("access_token", accessToken)
		q.Add("client_secret", config.VkontakteAppSecret)
		q.Add("token", vk.token)
		q.Add("v", "5.68")
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			ge = graceful.NewNetworkError(err.Error())
			break
		}
		err = json.NewDecoder(resp.Body).Decode(&f)
		resp.Body.Close() //defer replacement due to resource leak
		if err != nil {
			ge = graceful.NewParsingError(err.Error())
			break
		}
		if f.Error != nil {
			switch f.Error.Code {
			case 15:
				ge = graceful.NewNotFoundError(f.Error.Message, f.Error.Code)
			case 28: //обработка {"error":"d=vk; c=[28]; m=Application authorization failed: refresh service token"}
				serviceKey.Reset()
				fallthrough
			default:
				ge = graceful.NewVkError(f.Error.Message, f.Error.Code)
			}
		}
	}
	if ge != nil {
		return
	}
	if f.Response.Success != 1 {
		ge = graceful.NewNotFoundError("bad success flag")
		return
	}
	if f.Response.UserID == 0 {
		ge = graceful.NewInvalidError("empty user id")
		return
	}
	userID = fmt.Sprint(f.Response.UserID)
	return
}

func (vk VkToken) get(userID string) (string, *graceful.Error) {
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
	log.Debug("vk.get")
	req, err := http.NewRequest("GET", "https://api.vk.com/method/users.get", nil)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("fields", "sex,bdate,city,country")
	q.Add("user_ids", userID)
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
	if len(f.Response) != 1 || fmt.Sprint(f.Response[0].ID) != userID {
		return "", graceful.NewInvalidError("user id not match or empty response")
	}
	return f.Response[0].FirstName + " " + f.Response[0].LastName, nil
}

//GetUserInfo - method to check validity and get user information about the token if it valid. Returns NotFound error if token is not valid
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
