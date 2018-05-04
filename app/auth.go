package app

import (
	"errors"
	"gamelink-go/common"
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

var (
	checkAuthToken         = storage.CheckAuthToken
	checkRegister          = storage.CheckRegister
	generateStoreAuthToken = storage.GenerateStoreAuthToken
)

func (a *App) authMiddleware(ctx iris.Context) {
	log.Debug("app.authMiddleware")
	var status int
	var err error
	var userID int64
	token := ctx.GetHeader("Authorization")
	if token == "" {
		status = http.StatusUnauthorized
		err = errors.New("missing authorization header")
		goto sendErrorOrNext
	}
	if userID, err = checkAuthToken(token, a.redis); err != nil {
		switch err.(type) {
		case graceful.UnauthorizedError:
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
	var authToken string
	var thirdPartyToken common.IUserInfoGetter
	var userID int64
	var status = http.StatusOK
	var err error
	qs := ctx.Request().URL.Query()
	if vk, fb := qs["vk"], qs["fb"]; vk != nil && len(vk) == 1 && fb == nil {
		thirdPartyToken = social.VkToken(vk[0])
	} else if fb != nil && len(fb) == 1 && vk == nil {
		thirdPartyToken = social.FbToken(fb[0])
	} else {
		status = http.StatusBadRequest
		err = errors.New("query without vk or fb token")
		goto sendResponce
	}
	userID, err = checkRegister(thirdPartyToken, a.mySQL)
	if err != nil {
		switch err.(type) {
		case graceful.UnauthorizedError:
			status = http.StatusUnauthorized //пример использования супер домена ошибок "не найдено"
		default:
			status = http.StatusInternalServerError
		}
		goto sendResponce
	}
	authToken, err = generateStoreAuthToken(userID, a.redis)
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
