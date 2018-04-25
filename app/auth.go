package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/kataras/iris"
	"net/http"
	"gamelink-go/config"
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

func (a *App) authMiddleware(ctx iris.Context) {
	log.Debug("app.authMiddleware")
	var sendError = func(status int, err *graceful.Error) {
		log.WithError(err).WithFields(log.Fields{
			"remote": ctx.RemoteAddr(),
			"path":   ctx.Path(),
			"method": ctx.Method(),
		}).Error("check auth failed")
		ctx.ResponseWriter().WriteHeader(status)
		if config.IsDevelopmentEnv() && err != nil {
			ctx.JSON(J{"error": err.Error()})
		}
	}
	token := ctx.GetHeader("Authorization")
	if token == "" {
		sendError(http.StatusUnauthorized, graceful.NewInvalidError("missing authorization header"))
		return
	}
	if userId, err := storage.CheckAuthToken(token, a.Redis); err != nil {
		switch err.Domain() {
		case graceful.InvalidDomain:
			sendError(http.StatusUnauthorized, err)
		default:
			sendError(http.StatusInternalServerError, err)
		}
		return
	} else {
		ctx.Values().Set(userIdValueKey, userId)
		if config.IsDevelopmentEnv() {
			log.WithFields(log.Fields{
				"remote": ctx.RemoteAddr(),
				"path":   ctx.Path(),
				"method": ctx.Method(),
				"userId": userId,
			}).Info("new request")
		}
		ctx.Next()
		return
	}
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
		log.WithError(err).Error("get social user info failed")
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
		log.WithError(err).Error("db operations failed")
		status = http.StatusInternalServerError
		goto sendResponce
	}
	authToken, err = storage.GenerateStoreAuthToken(int64(userId), a.Redis) //TODO: фейковая конвертация для того чтобы собирался проект! убрать ее!
	if err != nil {
		log.WithError(err).Error("store token failed")
		status = http.StatusInternalServerError
		goto sendResponce
	}
	log.WithField("token", authToken).Debug("store token ok")
sendResponce:
	ctx.ResponseWriter().WriteHeader(status)
	if config.IsDevelopmentEnv() && err != nil {
		ctx.JSON(J{"error": err.Error()})
	} else if err == nil {
		ctx.JSON(J{"token": authToken})
	}
}