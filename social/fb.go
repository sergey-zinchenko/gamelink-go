package social

import (
	"net/http"
	"gamelink-go/config"
	"encoding/json"
	"gamelink-go/graceful"
	"net/url"
	log "github.com/sirupsen/logrus"
	"path"
)

type (
	FbToken struct {
		token string
	}

	fbError struct {
		Message string `json:"message"`
		Code int `json:"code"`
	}
)

func NewFbToken(token string) *FbToken {
	return &FbToken{token}
}

func (fb FbToken) debugToken() (string, *graceful.Error) {
	type (
		fbDebugTokenData struct {
			IsValid bool `json:"is_valid"`
			AppId string `json:"app_id"`
			UserId string `json:"user_id"`
		}

		fbDebugTokenResponse struct {
			Data fbDebugTokenData `json:"data"`
			Error *fbError `json:"error"`
		}
	)
	log.Debug("fb.debugToken")
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/debug_token", nil)
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	q := req.URL.Query()
	q.Add("access_token", config.FaceBookAppId + "|" + config.FaceBookAppSecret)
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
		return "", graceful.NewFbError(f.Error.Message, f.Error.Code)
	}
	if !f.Data.IsValid {
		return "", graceful.NewInvalidError("bad is_valid flag", InvalidOrUnsuccessCode)
	}
	if f.Data.AppId != config.FaceBookAppId || f.Data.UserId == "" {
		return "", graceful.NewInvalidError("invalid response format app_id or user_id", WrongApplicationOrEmptyUserIdCode)
	}
	return f.Data.UserId, nil
}

func (fb FbToken) get(userId string) (string, *graceful.Error) {
	type (
		fbGetResponse struct {
			Name string `json:"name"`
			Id string `json:"id"`
			Error *fbError `json:"error"`
		}
	)
	log.Debug("fb.get")
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return "", graceful.NewNetworkError(err.Error())
	}
	u.Path = path.Join(u.Path, userId)
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
	if f.Id != userId {
		return "", graceful.NewInvalidError( "user id not match")
	}
	return f.Name, nil
}

func (fb FbToken) GetUserInfo() (string, string, *graceful.Error) {
	log.Debug("fb.GetUserInfo")
	id, err := fb.debugToken()
	if err != nil {
		return "", "", err
	}
	name, err:= fb.get(id)
	if err != nil {
		return id, "", err
	}
	return id, name, nil
}