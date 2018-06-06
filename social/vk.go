package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"gamelink-go/config"
	"gamelink-go/graceful"
	"net/http"
	"strconv"
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

	//VkIdentifier - class to store vk identifier and column name
	VkIdentifier string
	//VkInfo - class for VK user info
	VkInfo struct {
		VkID VkIdentifier `json:"vk_id"`
		commonInfo
	}
)

var serviceKey VkServiceKey

const (
	maxRetries = 10
	//VkID - const name of vkontakte id column in the db
	VkID = "vk_id"
)

//Name - vk column name in the db
func (i VkIdentifier) Name() string {
	return VkID
}

//Value - vk identifier value
func (i VkIdentifier) Value() string {
	return string(i)
}

//ID - return vk id
func (d VkInfo) ID() ThirdPartyID {
	return d.VkID
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
			Email            string `json:"email"`
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
	//email := f.Email
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

func (token VkToken) get(userID string) (string, string, int64, *string, error) {
	type (
		vkLocation struct {
			LocID   int64   `json:"id"`
			LocName *string `json:"title,omitempty"`
		}

		vkUsersGetData struct {
			ID        int64       `json:"id"`
			FirstName string      `json:"first_name"`
			LastName  string      `json:"last_name"`
			Bdate     string      `json:"bdate"`
			Sex       int64       `json:"sex"`
			Country   *vkLocation `json:"country,omitempty"`
		}

		vkUsersGetResponse struct {
			Response []vkUsersGetData `json:"response"`
			Error    *vkError         `json:"error"`
		}
	)
	req, err := http.NewRequest("GET", "https://api.vk.com/method/users.get", nil)
	if err != nil {
		return "", "", 0, nil, err
	}
	q := req.URL.Query()
	q.Add("fields", "sex,bdate, country")
	q.Add("access_token", string(token))
	q.Add("user_ids", userID)
	q.Add("lang", "en")
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, nil, err
	}
	defer resp.Body.Close()
	var f vkUsersGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", "", 0, nil, err
	}
	if f.Error != nil {
		return "", "", 0, nil, NewVkError(f.Error.Message, f.Error.Code)
	}
	if len(f.Response) != 1 || fmt.Sprint(f.Response[0].ID) != userID {
		return "", "", 0, nil, errors.New("user id not match or empty response")
	}
	var country *string
	if f.Response[0].Country.LocName != nil {
		country = f.Response[0].Country.LocName
	}

	return f.Response[0].FirstName + " " + f.Response[0].LastName, f.Response[0].Bdate, f.Response[0].Sex, country, nil
}

//UserInfo - method to check validity and get user information about the token if it valid. Returns NotFound error if token is not valid
func (token VkToken) UserInfo() (ThirdPartyUser, error) {
	if token == "" {
		return nil, graceful.UnauthorizedError{Message: "empty token"}
	}
	id, err := token.checkToken()
	if err != nil {
		return nil, err
	}
	name, bdate, sex, country, err := token.get(id)
	if err != nil {
		return nil, err
	}
	friendsIds, err := token.getFriends(id)
	if err != nil {
		friendsIds = nil
	}
	var userSex string
	if sex == 1 {
		userSex = "F"
	} else if sex == 2 {
		userSex = "M"
	} else {
		userSex = "X"
	}
	userInfo := VkInfo{VkIdentifier(id), commonInfo{name, bdate, userSex, "", country, friendsIds}}
	return userInfo, nil
}

func (token VkToken) getFriends(userID string) ([]ThirdPartyID, error) {
	type (
		vkFriendsGetResponse struct {
			Data  []int    `json:"response"`
			Error *vkError `json:"error"`
		}
	)
	req, err := http.NewRequest("GET", "https://api.vk.com/method/friends.getAppUsers", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("access_token", string(token))
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var f vkFriendsGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return nil, err
	}
	if f.Error != nil {
		return nil, NewVkError(f.Error.Message, f.Error.Code)
	}

	friendsIds := make([]ThirdPartyID, len(f.Data))
	for k := range friendsIds {
		friendsIds[k] = VkIdentifier(strconv.Itoa(f.Data[k]))
	}
	return friendsIds, nil
}
