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
	"time"
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

//IsDummy - return true if this user auth without social
func (d VkInfo) IsDummy() bool {
	return false
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

func (token VkToken) get(userInfo *VkInfo) error {
	type (
		vkLocation struct {
			LocID   int64   `json:"id"`
			LocName *string `json:"title,omitempty"`
		}

		vkUsersGetData struct {
			ID        int64       `json:"id"`
			FirstName *string     `json:"first_name"`
			LastName  *string     `json:"last_name"`
			Bdate     *string     `json:"bdate"`
			Sex       *int64      `json:"sex"`
			Country   *vkLocation `json:"country"`
		}

		vkUsersGetResponse struct {
			Response []vkUsersGetData `json:"response"`
			Error    *vkError         `json:"error"`
		}
	)
	if userInfo == nil {
		return errors.New("userInfo pointer is nil")
	}
	if userInfo.VkID == "" {
		return errors.New("userInfo VK id must be set")
	}
	req, err := http.NewRequest("GET", "https://api.vk.com/method/users.get", nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("fields", "sex,bdate, country")
	q.Add("access_token", string(token))
	q.Add("user_ids", userInfo.VkID.Value())
	q.Add("lang", "en")
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var f vkUsersGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return err
	}
	if f.Error != nil {
		return NewVkError(f.Error.Message, f.Error.Code)
	}
	if len(f.Response) != 1 || fmt.Sprint(f.Response[0].ID) != userInfo.VkID.Value() {
		return errors.New("user id not match or empty response")
	}

	if f.Response[0].FirstName != nil && f.Response[0].LastName != nil {
		userInfo.FullName = *f.Response[0].FirstName + " " + *f.Response[0].LastName
	} else {
		return errors.New("name or last name can not be blank")
	}

	if f.Response[0].Country != nil && f.Response[0].Country.LocName != nil {
		userInfo.UserCountry = *f.Response[0].Country.LocName
	}
	if f.Response[0].Bdate != nil {
		bdate := *f.Response[0].Bdate
		lay := "2.1.2006"
		t, err := time.Parse(lay, bdate)
		if err != nil {
			return errors.New("can't parse bdate")
		}
		userInfo.Bdate = t.Unix()
	}
	if f.Response[0].Sex != nil {
		if *f.Response[0].Sex == 1 {
			userInfo.UserSex = "F"
		} else if *f.Response[0].Sex == 2 {
			userInfo.UserSex = "M"
		}
	}

	return nil
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
	userInfo := VkInfo{VkIdentifier(id), commonInfo{}}
	err = token.get(&userInfo)
	if err != nil {
		return nil, err
	}
	err = token.getFriends(&userInfo)
	if err != nil {
		if _, ok := err.(VkError); !ok {
			return nil, err
		}
	}

	return userInfo, nil
}

func (token VkToken) getFriends(userInfo *VkInfo) error {
	type (
		vkFriendsGetResponse struct {
			Data  []int    `json:"response"`
			Error *vkError `json:"error"`
		}
	)
	req, err := http.NewRequest("GET", "https://api.vk.com/method/friends.getAppUsers", nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("access_token", string(token))
	q.Add("v", "5.68")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var f vkFriendsGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return err
	}
	if f.Error != nil {
		return NewVkError(f.Error.Message, f.Error.Code)
	}

	friendsIds := make([]ThirdPartyID, len(f.Data))
	for k := range friendsIds {
		friendsIds[k] = VkIdentifier(strconv.Itoa(f.Data[k]))
	}
	userInfo.friends = friendsIds

	return nil
}
