package app

import (
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
		err  error
		user *storage.User
	)
	defer func() {
		if err != nil {
			if sc, hasCode := err.(graceful.StatusCode); hasCode {
				ctx.StatusCode(sc.StatusCode())
			} else {
				ctx.StatusCode(http.StatusInternalServerError)
			}
			ctx.Values().Set(errorCtxKey, err)
			return
		}
		ctx.Next()
	}()
	token := ctx.GetHeader("Authorization")
	if token == "" {
		err = graceful.UnauthorizedError{Message: "authorization token not set"}
		return
	}
	user, err = a.dbs.AuthorizedUser(token)
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
		if err == nil {
			ctx.JSON(j{"token": authToken})
		} else {
			if sc, hasCode := err.(graceful.StatusCode); hasCode {
				ctx.StatusCode(sc.StatusCode())
			} else {
				ctx.StatusCode(http.StatusInternalServerError)
			}
			ctx.Values().Set(errorCtxKey, err)
		}
	}()
	thirdPartyToken := tokenFromValues(ctx.Request().URL.Query())
	if thirdPartyToken == nil {
		err = graceful.BadRequestError{Message: "query without vk or fb token"}
		return
	}
	user, err = a.dbs.ThirdPartyUser(thirdPartyToken)
	if err != nil {
		return
	}
	authToken, err = user.AuthToken()
	if err != nil {
		return
	}
}
