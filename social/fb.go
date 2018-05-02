package social

import (
	"encoding/json"
	"gamelink-go/config"
	"gamelink-go/graceful"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path"
)

type (
	//FbToken - Class to get information about Facebook user tokens
	FbToken struct {
		token string
	}

	fbError struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
)

//NewFbToken - FbToken constructor, actually does nothing
func NewFbToken(token string) *FbToken {
	return &FbToken{token}
}

func (fb FbToken) debugToken() (string, *graceful.Error) {
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
	log.Debug("fb.debugToken")
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/debug_token", nil)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("access_token", config.FaceBookAppID+"|"+config.FaceBookAppSecret)
	q.Add("input_token", fb.token)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	defer resp.Body.Close()
	var f fbDebugTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", graceful.NewParsingError(err.Error())
	}
	if f.Error != nil {
		switch f.Error.Code {
		case 102, 190:
			return "", graceful.NewNotFoundError(f.Error.Message, f.Error.Code)
		default:
			return "", graceful.NewFbError(f.Error.Message, f.Error.Code)
		}
	}
	if !f.Data.IsValid {
		return "", graceful.NewNotFoundError("bad is_valid flag")
	}
	if f.Data.AppID != config.FaceBookAppID || f.Data.UserID == "" {
		return "", graceful.NewInvalidError("invalid response format app_id or user_id")
	}
	return f.Data.UserID, nil
}

func (fb FbToken) get(userID string) (string, *graceful.Error) {
	type (
		fbGetResponse struct {
			Name  string   `json:"name"`
			ID    string   `json:"id"`
			Error *fbError `json:"error"`
		}
	)
	log.Debug("fb.get")
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	u.Path = path.Join(u.Path, userID)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("fields", "id, name")
	q.Add("access_token", fb.token)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	defer resp.Body.Close()
	var f fbGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", graceful.NewParsingError(err.Error())
	}
	if f.Error != nil {
		return "", graceful.NewFbError(f.Error.Message, f.Error.Code)
	}
	if f.ID != userID {
		return "", graceful.NewInvalidError("user id not match")
	}
	return f.Name, nil
}

//GetUserInfo - method to get user information (name and identifier) of a valid user token and returns error (d = NotFound) if invalid
func (fb FbToken) GetUserInfo() (string, string, *graceful.Error) {
	log.Debug("fb.GetUserInfo")
	id, err := fb.debugToken()
	if err != nil {
		return "", "", err
	}
	name, err := fb.get(id)
	if err != nil {
		return id, "", err
	}
	return id, name, nil
}
