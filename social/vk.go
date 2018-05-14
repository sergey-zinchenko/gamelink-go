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
	VkToken string

	//VkServiceKey - structure for store and request vk service key
	VkServiceKey struct {
		key string
		m   sync.Mutex
	}

	vkIdentifier string
)

var serviceKey VkServiceKey

const (
	maxRetries = 10
)

func (i vkIdentifier) Name() string {
	return "vk_id"
}

func (i vkIdentifier) Value() string {
	return string(i)
}

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
		return "", NewVkError(f.ErrorDescription)
	}
	if f.AccessToken == "" {
		return "", errors.New("empty access_token")
	}
	sk.key = f.AccessToken
	return sk.key, nil
}

//Obsolete - method removes stored service key if it is the same as parameter and new one will be requested from server on next Key() call
func (sk *VkServiceKey) Obsolete(old string) {
	sk.m.Lock()
	defer sk.m.Unlock()
	if sk.key == old {
		sk.key = ""
	}
}

func (token VkToken) checkToken() (userID string, err error) {
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
		q.Add("token", string(token))
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
				err = &graceful.UnauthorizedError{Message: fmt.Sprintf("%d:%s", f.Error.Code, f.Error.Message)}
			case 28: //обработка {"error":"d=vk; c=[28]; m=Application authorization failed: refresh service token"}
				serviceKey.Obsolete(accessToken)
				fallthrough
			default:
				err = NewVkError(f.Error.Message, f.Error.Code)
			}
		}
	}
	if err != nil {
		return
	}
	if f.Response.Success != 1 {
		err = &graceful.UnauthorizedError{Message: "incorrect success flag"}
		return
	}
	if f.Response.UserID == 0 {
		err = errors.New("empty user id")
		return
	}
	userID = fmt.Sprint(f.Response.UserID)
	return
}

func (token VkToken) get(userID string) (string, error) {
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
	q.Add("access_token", string(token))
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
		return "", NewVkError(f.Error.Message, f.Error.Code)
	}
	if len(f.Response) != 1 || fmt.Sprint(f.Response[0].ID) != userID {
		return "", errors.New("user id not match or empty response")
	}
	return f.Response[0].FirstName + " " + f.Response[0].LastName, nil
}

//UserInfo - method to check validity and get user information about the token if it valid. Returns NotFound error if token is not valid
func (token VkToken) UserInfo() (ThirdPartyID, string, []string, error) {
	if token == "" {
		return nil, "", nil, graceful.UnauthorizedError{Message: "empty token"}
	}
	id, err := token.checkToken()
	if err != nil {
		return nil, "", nil, err
	}
	name, err := token.get(id)
	if err != nil {
		return vkIdentifier(id), "", nil, err
	}
	return vkIdentifier(id), name, nil, nil
}
