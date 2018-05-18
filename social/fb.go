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

func (token FbToken) get(userID string) (string, []ThirdPartyID, string, string, string, error) {
	type (
		fbFriends struct {
			FbFriendID string `json:"id"`
		}

		fbFriendsData struct {
			Data []*fbFriends
		}

		fbGetResponse struct {
			Name    string         `json:"name"`
			ID      string         `json:"id"`
			Friends *fbFriendsData `json:"friends"`
			Sex     string         `json:"gender"`
			Bdate   string         `json:"birthday"`
			Email   string         `json:"email"`
			Error   *fbError       `json:"error"`
		}
	)
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return "", nil, "", "", "", err
	}
	u.Path = path.Join(u.Path, userID)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", nil, "", "", "", err
	}
	q := req.URL.Query()
	q.Add("fields", "id, name, friends,gender,birthday,email")
	q.Add("access_token", string(token))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, "", "", "", err
	}
	defer resp.Body.Close()
	var f fbGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", nil, "", "", "", err
	}
	if f.Error != nil {
		return "", nil, "", "", "", NewFbError(f.Error.Message, f.Error.Code)
	}
	if f.ID != userID {
		return "", nil, "", "", "", errors.New("user id not match")
	}
	friendsIds := make([]ThirdPartyID, len(f.Friends.Data))
	for k := range friendsIds {
		friendsIds[k] = FbIdentifier(f.Friends.Data[k].FbFriendID)
	}

	return f.Name, friendsIds, f.Sex, f.Bdate, f.Email, nil
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
	name, friendsIds, sex, bdate, email, err := token.get(id)
	if err != nil {
		return nil, err
	}
	var userSex string
	if sex == "male" {
		userSex = "M"
	} else if sex == "female" {
		userSex = "F"
	} else {
		userSex = "X"
	}
	info := commonInfo{name, bdate, userSex, email, friendsIds}
	userInfo := FbInfo{FbIdentifier(id), info}
	return userInfo, nil
}
