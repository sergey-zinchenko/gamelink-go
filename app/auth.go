package app

import (
	"errors"
	"gamelink-go/graceful"
	"gamelink-go/social"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"net/http"
	"net/url"
)

const (
	userCtxKey = "user"
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
		status int
		err    error
		user   *storage.User
	)
	defer func() {
		if err != nil {
			ctx.StatusCode(status)
			ctx.Values().Set(errorCtxKey, err)
			return
		}
		ctx.Next()
	}()
	token := ctx.GetHeader("Authorization")
	if token == "" {
		status = http.StatusUnauthorized
		err = errors.New("missing authorization header")
		return
	}
	user, err = a.dbs.AuthorizedUser(token)
	if err != nil {
		switch err.(type) {
		case graceful.UnauthorizedError:
			status = http.StatusUnauthorized
		default:
			status = http.StatusInternalServerError
		}
		return
	}
	ctx.Values().Set(userCtxKey, user)
}

func (a *App) registerLogin(ctx iris.Context) {
	var (
		authToken string
		user      *storage.User
		status    = http.StatusOK
		err       error
	)
	defer func() {
		ctx.StatusCode(status)
		if err == nil {
			ctx.JSON(j{"token": authToken})
		} else {
			ctx.Values().Set(errorCtxKey, err)
		}
	}()
	thirdPartyToken := tokenFromValues(ctx.Request().URL.Query())
	if thirdPartyToken == nil {
		status = http.StatusBadRequest
		err = errors.New("query without vk or fb token")
		return
	}
	user, err = a.dbs.ThirdPartyUser(thirdPartyToken)
	if err != nil {
		switch err.(type) {
		case graceful.UnauthorizedError:
			status = http.StatusUnauthorized //пример использования супер домена ошибок "не найдено"
		default:
			status = http.StatusInternalServerError
		}
		return
	}
	authToken, err = user.AuthToken()
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
}
