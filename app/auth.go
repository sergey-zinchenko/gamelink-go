package app

import (
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

const (
	userCtxKey        = "user"
	firebaseMsgSystem = "firebase"
	apnsMsgSystem     = "apns"
)

var (
	tokenFromValues = func(query url.Values) social.ThirdPartyToken {
		if vk, fb := query["vk"], query["fb"]; vk != nil && len(vk) == 1 && fb == nil {
			return social.VkToken(vk[0])
		} else if fb != nil && len(fb) == 1 && vk == nil {
			return social.FbToken(fb[0])
		}
		return nil
	}
)

func (a *App) authMiddleware(ctx iris.Context) {
	var (
		err  error
		user *storage.User
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
			ctx.StopExecution()
		}
		ctx.Next()
	}()
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	arr := strings.Split(header, " ")
	if len(arr) < 2 {
		err = graceful.BadRequestError{Message: "authorization header not valid"}
		return
	}
	if strings.ToUpper(arr[0]) != "BEARER" {
		err = graceful.BadRequestError{Message: "authorization header not valid"}
		return
	}
	user, err = a.dbs.AuthorizedUser(arr[1])
	if err != nil {
		return
	}
	err = user.AddDeviceID(a.checkDeviceHeader(ctx))
	if err != nil {
		return
	}
	ctx.Values().Set(userCtxKey, user)
}

func (a *App) registerLogin(ctx iris.Context) {
	var (
		authToken string
		user      *storage.User
		err       error
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	thirdPartyToken := tokenFromValues(ctx.Request().URL.Query())
	if thirdPartyToken == nil {
		thirdPartyToken = social.NewDummyToken()
	}
	user, err = a.dbs.ThirdPartyUser(thirdPartyToken)
	if err != nil {
		logrus.Warn(err.Error())
		return
	}
	authToken, err = user.AuthToken(thirdPartyToken.IsDummy())
	if err != nil {
		logrus.Warn(err.Error())
		return
	}
	err = user.AddDeviceID(a.checkDeviceHeader(ctx))
	if err != nil {
		logrus.Warn(err.Error())
		return
	}
	ctx.JSON(C.J{"token": authToken})
}

func (a *App) checkDeviceHeader(ctx iris.Context) (string, string) {
	firebaseHeader := ctx.GetHeader(firebaseMsgSystem)
	if firebaseHeader != "" {
		return firebaseHeader, firebaseMsgSystem
	}
	apnsHeader := ctx.GetHeader(apnsMsgSystem)
	if apnsHeader != "" {
		return apnsHeader, apnsMsgSystem
	}
	return "", ""
}
