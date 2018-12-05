package app

import (
	"fmt"
	C "gamelink-go/common"
	"gamelink-go/config"
	"gamelink-go/graceful"
	push "gamelink-go/proto_nats_msg"
	"gamelink-go/storage"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func (a *App) getUser(ctx iris.Context) {
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err := user.DataString()
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	if err != nil {
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	_, err = ctx.WriteString(data)
}

func (a *App) postUser(ctx iris.Context) {
	var (
		data, updated C.J
		newScore      int
		receivers     []*push.UserInfo
		err           error
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	err = ctx.ReadJSON(&data)
	if err != nil {
		return
	}
	if config.PushWhenOutrun {
		for lbNum := 1; lbNum < storage.NumOfLeaderBoards+1; lbNum++ {
			if data[fmt.Sprintf("lb%d", lbNum)] != nil {
				newScore = int(data[fmt.Sprintf("lb%d", lbNum)].(float64))
				r, err := user.GetPushReceivers(newScore, lbNum)
				if err != nil {
					logrus.Warn(err.Error()) //Тут можно бы вставить и ретурны, но нужно валить сохранение юезра, если косяк с пушами? К юзеру то это не имеет отношения
				}
				receivers = append(receivers, r...)
			}
		}
	}
	updated, err = user.Update(data)
	if err != nil {
		return
	}
	if receivers != nil {
		var userName string
		if updated["nickname"] != nil {
			userName = updated["nickname"].(string)
		} else {
			userName = updated["name"].(string)
		}
		msg := fmt.Sprintf("Hey! Your friend %s beat you. Check the leaderboard to make sure", userName)
		err = a.nc.PrepareAndPushMessage(msg, receivers)
		if err != nil {
			logrus.Warn(err.Error()) //Тут можно бы вставить и ретурны, но нужно валить сохранение юезра, если косяк с пушами? К юзеру то это не имеет отношения
		}
	}
	_, err = ctx.JSON(updated)
}

func (a *App) deleteUser(ctx iris.Context) {
	var (
		err  error
		data C.J
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	data, err = user.Delete(ctx.Request().URL.Query()["fields"])
	if err != nil {
		return
	}
	if data == nil {
		ctx.StatusCode(http.StatusNoContent)
	} else {
		_, err = ctx.JSON(data)
	}
}

func (a *App) addAuth(ctx iris.Context) {
	var (
		err       error
		data      C.J
		existedID int64
	)
	defer func() {
		if err != nil {
			handleError(err, ctx)
		}
	}()
	user := ctx.Values().Get(userCtxKey).(*storage.User)
	token := tokenFromValues(ctx.Request().URL.Query())
	if token == nil {
		err = graceful.BadRequestError{Message: "invalid token"}
		return
	}
	data, existedID, err = user.AddSocial(token)
	if err != nil {
		return
	}
	response := make(map[string]interface{})
	response["data"] = data
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	arr := strings.Split(header, " ")
	tokenValue := arr[1]
	if tokenValue != "" && tokenValue[:5] == "dummy" {
		var updID int64
		if existedID != 0 {
			updID = existedID
		} else {
			updID = user.ID()
		}
		newToken, err := a.dbs.AuthToken(false, updID) //false cause we want generate new token for normal user not for dummy
		if err != nil {
			return
		}
		err = a.dbs.DeleteRedisToken(tokenValue)
		if err != nil {
			return
		}
		response["token"] = newToken
	}
	_, err = ctx.JSON(response)
}
