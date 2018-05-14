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
	//FbFriends - Class to get information about Facebook users friends who install this app
	FbFriends struct {
		FbFriendID string `json:"id"`
	}
	//FbFriendsData - friends array
	FbFriendsData struct {
		Data []*FbFriends
	}

	fbError struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	fbIdentifier string
)

func (i fbIdentifier) Name() string {
	return "fb_id"
}

func (i fbIdentifier) Value() string {
	return string(i)
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

func (token FbToken) get(userID string) (string, []string, error) {
	type (
		fbGetResponse struct {
			Name    string         `json:"name"`
			ID      string         `json:"id"`
			Friends *FbFriendsData `json:"friends"`
			Error   *fbError       `json:"error"`
		}
	)
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return "", nil, err
	}
	u.Path = path.Join(u.Path, userID)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", nil, err
	}
	q := req.URL.Query()
	q.Add("fields", "id, name, friends")
	q.Add("access_token", string(token))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	var f fbGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", nil, err
	}
	if f.Error != nil {
		return "", nil, NewFbError(f.Error.Message, f.Error.Code)
	}
	if f.ID != userID {
		return "", nil, errors.New("user id not match")
	}
	var friendsIds = prepareUserFriendsArray(f.Friends.Data)
	return f.Name, friendsIds, nil
}

//UserInfo - method to get user information (name and identifier) of a valid user token and returns error (d = NotFound) if invalid
func (token FbToken) UserInfo() (ThirdPartyID, string, []string, error) {
	if token == "" {
		return nil, "", nil, graceful.UnauthorizedError{Message: "empty token"}
	}
	id, err := token.debugToken()
	if err != nil {
		return nil, "", nil, err
	}
	name, friendsIds, err := token.get(id)
	if err != nil {
		return fbIdentifier(id), "", nil, err
	}
	return fbIdentifier(id), name, friendsIds, nil
}

func prepareUserFriendsArray(friends []*FbFriends) []string {
	if len(friends) == 0 {
		return nil
	}
	var friendsIds = make([]string, len(friends))
	for k := range friends {
		friendsIds[k] = friends[k].FbFriendID
	}
	return friendsIds
}
