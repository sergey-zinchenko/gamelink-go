package social

import (
	"net/http"
	"gamelink-go/config"
	"encoding/json"
	"gamelink-go/graceful"
	"net/url"
	"path"
	log "github.com/sirupsen/logrus"
)

type (
	FbToken struct {
		token string
	}
)

func NewFbToken(token string) *FbToken {
	return &FbToken{token}
}

func (fb FbToken) debugToken() (string, *graceful.Error) {
	type (
		fbDebugTokenData struct {
			Is_valid bool
			App_id string
			User_id string
		}

		fbDebugTokenResponse struct {
			Data fbDebugTokenData
		}
	)
	log.Debug("fb.debugToken")
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v2.8/debug_token", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("access_token", config.FaceBookAppId + "|" + config.FaceBookAppSecret)
	q.Add("input_token", fb.token)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var f fbDebugTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return
	}
	if f.Data.App_id != config.FaceBookAppId || !f.Data.Is_valid || f.Data.User_id == "" {
		err = errors.New("invalid response format app_id or is_valid or user_id")
		return
	}
	userId = f.Data.User_id
	return
}

func (fb FbToken) get(userId string) (string, *graceful.Error) {
	type (
		fbGetResponse struct {
			Name string
			Id string
		}
	)
	log.Debug("fb.get")
	u, err := url.Parse("https://graph.facebook.com/v2.8")
	if err != nil {
		return
	}
	u.Path = path.Join(u.Path, userId)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("fields", "id, name")
	q.Add("access_token", fb.token)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var f fbGetResponse
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		return
	}
	if f.Id != userId {
		err = errors.New("invalid response user id")
		return
	}
	name = f.Name
	return
}

func (fb FbToken) GetUserInfo() (string, string, *graceful.Error) {
	log.Debug("fb.GetUserInfo")
	err, id = fb.debugToken()
	if err != nil {
		return
	}
	err, name = fb.get(id)
	return
}
