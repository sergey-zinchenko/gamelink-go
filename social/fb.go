package social

import (
	"encoding/json"
	"errors"
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
)

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
			return "", &graceful.UnauthorizedError{}
		default:
			return "", NewFbError(f.Error.Message, f.Error.Code)
		}
	}
	if !f.Data.IsValid {
		return "", &graceful.UnauthorizedError{}
	}
	if f.Data.AppID != config.FaceBookAppID || f.Data.UserID == "" {
		return "", errors.New("invalid response format app_id or user_id")
	}
	return f.Data.UserID, nil
}

func (token FbToken) get(userID string) (string, error) {
	type (
		fbGetResponse struct {
			Name  string   `json:"name"`
			ID    string   `json:"id"`
			Error *fbError `json:"error"`
		}
	)
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, userID)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("fields", "id, name")
	q.Add("access_token", string(token))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var f fbGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return "", err
	}
	if f.Error != nil {
		return "", NewFbError(f.Error.Message, f.Error.Code)
	}
	if f.ID != userID {
		return "", errors.New("user id not match")
	}
	return f.Name, nil
}

//UserInfo - method to get user information (name and identifier) of a valid user token and returns error (d = NotFound) if invalid
func (token FbToken) UserInfo() (string, string, error) {
	if token == "" {
		return "", "", graceful.UnauthorizedError{}
	}
	id, err := token.debugToken()
	if err != nil {
		return "", "", err
	}
	name, err := token.get(id)
	if err != nil {
		return id, "", err
	}
	return id, name, nil
}
