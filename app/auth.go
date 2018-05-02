package app

import (
	"gamelink-go/graceful"
	"gamelink-go/social"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	userIDCtxKey = "userId"
)

func (a *App) authMiddleware(ctx iris.Context) {
	log.Debug("app.authMiddleware")
	var status int
	var err *graceful.Error
	var userID int64
	token := ctx.GetHeader("Authorization")
	if token == "" {
		status = http.StatusUnauthorized
		err = graceful.NewInvalidError("missing authorization header")
		goto sendErrorOrNext
	}
	if userID, err = storage.CheckAuthToken(token, a.Redis); err != nil {
		switch err.Domain() {
		case graceful.NotFoundDomain:
			status = http.StatusUnauthorized
		default:
			status = http.StatusInternalServerError
		}
		goto sendErrorOrNext
	} else {
		log.WithFields(log.Fields{"remote": ctx.RemoteAddr(),
			"user_id": userID}).Debug("authorized")
		ctx.Values().Set(userIDCtxKey, userID)
	}
sendErrorOrNext:
	if err != nil {
		ctx.StatusCode(status)
		ctx.Values().Set(errorCtxKey, err)
		return
	}
	ctx.Next()
}

func (a *App) registerLogin(ctx iris.Context) {
	log.Debug("app.registerLogin")
	var socialID, name, token, authToken string
	var userID int64
	var tokenSource social.TokenSource
	var status = http.StatusOK
	var err *graceful.Error
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
	socialID, name, err = social.GetSocialUserInfo(tokenSource, token)
	if err != nil {
		switch err.Domain() {
		case graceful.NotFoundDomain:
			status = http.StatusUnauthorized //пример использования супер домена ошибок "не найдено"
		default:
			status = http.StatusInternalServerError
		}
		goto sendResponce
	}
	userID, err = storage.CheckRegister(tokenSource, socialID, name, a.MySQL)
	if err != nil {
		status = http.StatusInternalServerError
		goto sendResponce
	}
	authToken, err = storage.GenerateStoreAuthToken(userID, a.Redis)
	if err != nil {
		status = http.StatusInternalServerError
		goto sendResponce
	}
	log.WithField("token", authToken).Debug("register or login ok")
sendResponce:
	ctx.StatusCode(status)
	if err == nil {
		ctx.JSON(j{"token": authToken})
	} else {
		ctx.Values().Set(errorCtxKey, err)
	}
}
