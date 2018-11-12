package app

import (
	C "gamelink-go/common"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"net/url"
	"reflect"
	"strings"
)

const (
	userCtxKey        = "user"
	iosDeviceType     = "ios"
	androidDeviceType = "android"
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
		err        error
		deviceID   string
		deviceType string
		user       *storage.User
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
	deviceID, deviceType = a.checkDeviceHeader(ctx)
	if deviceID != "" {
		err := user.AddDeviceID(deviceID, deviceType)
		if err != nil {
			return
		}
	}
	ctx.Values().Set(userCtxKey, user)
}

func (a *App) registerLogin(ctx iris.Context) {
	var (
		authToken string
		user      *storage.User
		//deviceID   string
		//deviceType string
		err error
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()

	//thirdPartyToken := tokenFromValues(ctx.Request().URL.Query())
	//if thirdPartyToken == nil {
	//	user, err = a.dbs.ThirdPartyUser(social.DummyToken(""))
	//	if err != nil {
	//		return
	//	}
	//	authToken, err = user.AuthToken(true)
	//} else {
	//	user, err = a.dbs.ThirdPartyUser(thirdPartyToken)
	//	if err != nil {
	//		return
	//	}
	//	authToken, err = user.AuthToken(false)
	//}

	//TODO: так же короче писать???
	thirdPartyToken := tokenFromValues(ctx.Request().URL.Query())
	if thirdPartyToken == nil {
		thirdPartyToken = social.DummyToken("")
	}
	authToken, err = user.AuthToken(reflect.TypeOf(thirdPartyToken) == reflect.TypeOf(social.DummyToken("")))
	//TODO: из за предыдущей строчки веротяно лучше сделать isDummy в токене - не красивая запись получается
	//deviceID, deviceType = a.checkDeviceHeader(ctx)
	//if deviceID != "" {
	//	err := user.AddDeviceID(deviceID, deviceType)
	//	if err != nil {
	//		return
	//	}
	//}

	//TODO:И тут таже история, естетсвенно функцию чуть допилить придется
	err = user.AddDeviceID(a.checkDeviceHeader(ctx))
	if err != nil {
		return
	}

	ctx.JSON(C.J{"token": authToken})
}

func (a *App) checkDeviceHeader(ctx iris.Context) (string, string) {
	if ctx.GetHeader("iosdevice") != "" {
		return ctx.GetHeader("iosdevice"), iosDeviceType
	}
	if ctx.GetHeader("androiddevice") != "" {
		return ctx.GetHeader("androiddevice"), androidDeviceType
	}
	return "", ""
}
