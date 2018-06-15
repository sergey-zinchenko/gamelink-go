package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"gamelink-go/config"
	"gamelink-go/graceful"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type (
	//FbToken - Class to get information about Facebook user tokens
	FbToken string

	fbError struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	//FbIdentifier - class to store fb identifier and column name
	FbIdentifier string
	//FbInfo - class to user info from FB
	FbInfo struct {
		FbID FbIdentifier `json:"fb_id"`
		commonInfo
	}
)

const (
	//FbID - const name of facebook identifier column in the db
	FbID = "fb_id"
)

//Name - fb column name in the db
func (i FbIdentifier) Name() string {
	return FbID
}

//Value - fb id value
func (i FbIdentifier) Value() string {
	return string(i)
}

//ID - return fbID
func (d FbInfo) ID() ThirdPartyID {
	return d.FbID
}

func (token FbToken) debugToken() (string, error) {
	type (
		fbDebugTokenData struct {
			IsValid bool   `json:"is_valid"`
			AppID   string `json:"app_id"`
			UserID  string `json:"user_id"`
		}

		fbDebugTokenResponse struct {
			Data  fbDebugTokenData `json:"data"`
			Error *fbError         `json:"error"`
		}
	)
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/debug_token", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("access_token", config.FaceBookAppID+"|"+config.FaceBookAppSecret)
	q.Add("input_token", string(token))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var f fbDebugTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", err
	}
	if f.Error != nil {
		switch f.Error.Code {
		case 102, 190:
			return "", &graceful.UnauthorizedError{Message: fmt.Sprintf("%d:%s", f.Error.Code, f.Error.Message)}
		default:
			return "", NewFbError(f.Error.Message, f.Error.Code)
		}
	}
	if !f.Data.IsValid {
		return "", &graceful.UnauthorizedError{Message: "wrong is_valid flag"}
	}
	if f.Data.AppID != config.FaceBookAppID || f.Data.UserID == "" {
		return "", errors.New("invalid response format app_id or user_id")
	}
	return f.Data.UserID, nil
}

func (token FbToken) get(userInfo *FbInfo) error {
	type (
		fbFriends struct {
			FbFriendID string `json:"id"`
		}

		fbFriendsData struct {
			Data []*fbFriends
		}

		fbLocInfo struct {
			LocName string  `json:"city"`
			Country *string `json:"country"`
		}

		fbLocation struct {
			LocInfo *fbLocInfo `json:"location"`
		}

		fbGetResponse struct {
			Name     *string        `json:"name"`
			ID       string         `json:"id"`
			Friends  *fbFriendsData `json:"friends"`
			Sex      *string        `json:"gender"`
			Bdate    *string        `json:"birthday"`
			Email    *string        `json:"email"`
			Location *fbLocation    `json:"location"`
			Error    *fbError       `json:"error"`
		}
	)
	if userInfo == nil {
		return errors.New("userInfo pointer is nil")
	}
	if userInfo.FbID == ""{
		errors.New("userInfo Fb ID must be set")
	}
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, userInfo.FbID.Value())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("fields", "id, name, friends,gender,birthday,email,location{location}")
	q.Add("locale", "en_GB")
	q.Add("access_token", string(token))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var f fbGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)

	if err != nil {
		return err
	}
	if f.Error != nil {
		return NewFbError(f.Error.Message, f.Error.Code)
	}
	if f.ID != userInfo.FbID.Value() {
		return errors.New("user id not match")
	}
	if f.Name != nil {
		userInfo.FullName = *f.Name
	} else {
		return errors.New("fb API return blank name")
	}

	if f.Bdate != nil {
		userInfo.Bdate = strings.Replace(*f.Bdate, "/", ".", 3)
	}
	if f.Sex != nil {
		if *f.Sex == "male" {
			userInfo.Sex = "M"
		} else if *f.Sex == "female" {
			userInfo.Sex = "F"
		} else {
			userInfo.Sex = "X"
		}
	}
	if f.Email != nil {
		userInfo.UserEmail = *f.Email
	}
	if f.Location != nil && f.Location.LocInfo.Country != nil {
		userInfo.UserCountry = *f.Location.LocInfo.Country
	}

	if f.Friends != nil && f.Friends.Data != nil{
	friendsIds := make([]ThirdPartyID, len(f.Friends.Data))
	for k := range friendsIds {
		friendsIds[k] = FbIdentifier(f.Friends.Data[k].FbFriendID)
	}
		userInfo.friends = friendsIds
	}
	return nil
}

//UserInfo - method to get user information (name and identifier) of a valid user token and returns error (d = NotFound) if invalid
func (token FbToken) UserInfo() (ThirdPartyUser, error) {
	if token == "" {
		return nil, graceful.UnauthorizedError{Message: "empty token"}
	}
	id, err := token.debugToken()
	if err != nil {
		return nil, err
	}
	userInfo := FbInfo{FbIdentifier(id), commonInfo{}}
	err = token.get(&userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}
