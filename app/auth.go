package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/kataras/iris"
	"net/http"
	"gamelink-go/config"
	"gamelink-go/graceful"
	"gamelink-go/storage"
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

func (a *App) registerLogin(ctx iris.Context) {
	var sendError = func(status int, err *graceful.Error, ctx iris.Context) {
		log.Debug("register login failed")
		ctx.ResponseWriter().WriteHeader(status)
		if config.IsDevelopmentEnv() && err != nil {
			ctx.JSON(J{"error": err.Error()})
		}
	}
	log.Debug("app.registerLogin")
	qs := ctx.Request().URL.Query()
	vk := qs["vk"]
	fb := qs["fb"]
	var userId int64
	var err *graceful.Error
	if vk != nil && len(vk) == 1 && fb == nil {
		log.WithField("vk_token", vk[0]).Debug("token received")
		userId, err = storage.VkCheckRegister(vk[0], a.MySql)
		if err != nil {
			var status int
			switch err.Domain() {
			case graceful.VkDomain:
				if hasCode, code := err.Code(); !hasCode {
					status = http.StatusInternalServerError
				} else {
					switch code {
					case 15:
						status = http.StatusUnauthorized
					default:
						status = http.StatusInternalServerError
					}
				}
			default:
				status = http.StatusInternalServerError
			}
			sendError(status, err, ctx)
			return
		}
	} else if fb != nil && len(fb) == 1 && vk == nil {
		log.WithField("fb_token", fb[0]).Debug("token received")
		userId, err = storage.FbCheckRegister(fb[0], a.MySql)
		if err != nil {
			var status int
			switch err.Domain() {
			case graceful.FbDomain:
				if hasCode, code := err.Code(); !hasCode {
					status = http.StatusInternalServerError
				} else {
					switch code {
					case 102, 190:
						status = http.StatusUnauthorized
					default:
						status = http.StatusInternalServerError
					}
				}
			case graceful.InvalidDomain:
				status = http.StatusUnauthorized
			default:
				status = http.StatusInternalServerError
			}
			sendError(status, err, ctx)
			return
		}
	} else {
		sendError(http.StatusBadRequest, nil, ctx)
		return
	}
	authToken, err := storage.GenerateStoreAuthToken(userId, a.Redis)
	if err != nil {
		log.WithError(err).Error("store token failed")
		sendError(http.StatusInternalServerError, err, ctx)
		return
	}
	log.WithField("token", authToken).Debug("store token ok")
	ctx.ResponseWriter().WriteHeader(http.StatusOK)
	ctx.JSON(J{"token": authToken})
}