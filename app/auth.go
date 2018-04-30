package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/kataras/iris"
	"net/http"
	"gamelink-go/graceful"
	"gamelink-go/storage"
	"gamelink-go/social"
)

type (
	J map[string]interface{}
)

const (
	userIdValueKey = "userId"
)

func (a *App) authMiddleware2(ctx iris.Context) {
	log.Debug("app.authMiddleware2")
	var status = http.StatusUnauthorized
	var err *graceful.Error
	var userId int64
	token := ctx.GetHeader("Authorization")
	if token == "" {
		status = http.StatusUnauthorized
		err = graceful.NewInvalidError("missing authorization header")
		goto sendErrorOrNext
	}
	if userId, err = storage.CheckAuthToken(token, a.Redis); err != nil {
		switch err.Domain() {
		case graceful.NotFoundDomain:
			status = http.StatusUnauthorized
		default:
			status = http.StatusInternalServerError
		}
		goto sendErrorOrNext
	} else {
		log.WithFields(log.Fields{"remote": ctx.RemoteAddr(),
			"user_id": userId}).Debug("authorized")
		ctx.Values().Set(userIdValueKey, userId)
	}
sendErrorOrNext:
	if err != nil {
		ctx.StatusCode(status)
		ctx.Values().Set("error", err)
		return
	}
	ctx.Next()
}

func (a *App) registerLogin2(ctx iris.Context) {
	log.Debug("app.registerLogin2")
	var socialId, name, token, authToken string
	var userId int64
	var tokenSource social.TokenSource
	var status = http.StatusOK
	var err *graceful.Error = nil
	qs := ctx.Request().URL.Query()
	if vk, fb := qs["vk"], qs["fb"]; vk != nil && len(vk) == 1 && fb == nil {
		token = vk[0]
		tokenSource = social.VKSource
	} else if fb != nil && len(fb) == 1 && vk == nil {
		token = fb[0]
		tokenSource = social.FbSource
	} else {
		status = http.StatusBadRequest
		err = graceful.NewInvalidError("query without vk or fb token")
		goto sendResponce
	}
	socialId, name, err = social.GetSocialUserInfo(tokenSource, token)
	if err != nil {
		switch err.Domain() {
		case graceful.NotFoundDomain:
			status = http.StatusUnauthorized //пример использования супер домена ошибок "не найдено"
		default:
			status = http.StatusInternalServerError
		}
		goto sendResponce
	}
	userId, err = storage.CheckRegister(tokenSource, socialId, name, a.MySql)
	if err != nil {
		status = http.StatusInternalServerError
		goto sendResponce
	}
	authToken, err = storage.GenerateStoreAuthToken(userId, a.Redis)
	if err != nil {
		status = http.StatusInternalServerError
		goto sendResponce
	}
	log.WithField("token", authToken).Debug("register or login ok")
sendResponce:
	ctx.StatusCode(status)
	if err == nil {
		ctx.JSON(J{"token": authToken})
	} else {
		ctx.Values().Set("error", err)
	}
}