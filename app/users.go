package app

import (
	"fmt"
	C "gamelink-go/common"
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
	if err != nil {
		handleError(err, ctx)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(data)
}

func (a *App) postUser(ctx iris.Context) {
	var (
		data, updated C.J
		newScore      int
		recievers     []*push.UserInfo
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
	for lbNum := 1; lbNum < storage.NumOfLeaderBoards+1; lbNum++ {
		if data[fmt.Sprintf("lb%d", lbNum)] != nil {
			newScore = int(data[fmt.Sprintf("lb%d", lbNum)].(float64))
			logrus.Warn(fmt.Sprintf("lb%d - ", lbNum), newScore)
			r, err := user.GetPushReceivers(newScore, lbNum)
			if err != nil {
				logrus.Warn(err.Error()) //Тут можно бы вставить и ретурны, но нужно валить сохранение юезра, если косяк с пушами? К юзеру то это не имеет отношения
			}
			logrus.Warn("get recievers", r)
			recievers = append(recievers, r...)
		}
	}
	updated, err = user.Update(data)
	if err != nil {
		return
	}
	if recievers != nil {
		var userName string
		if updated["username"] != nil {
			userName = updated["username"].(string)
		} else {
			userName = updated["name"].(string)
		}
		logrus.Warn("get username", userName)
		msg := fmt.Sprintf("Hey! Your friend %s beat you. Check the leaderboard to make sure", userName)
		err = a.nc.PrepareAndPushMessage(msg, recievers)
		if err != nil {
			logrus.Warn(err.Error()) //Тут можно бы вставить и ретурны, но нужно валить сохранение юезра, если косяк с пушами? К юзеру то это не имеет отношения
		}
	}
	ctx.JSON(updated)
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
		ctx.JSON(data)
	}
}

func (a *App) addAuth(ctx iris.Context) {
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
	token := tokenFromValues(ctx.Request().URL.Query())
	if token == nil {
		err = graceful.BadRequestError{Message: "invalid token"}
		return
	}
	data, err = user.AddSocial(token)
	if err != nil {
		return
	}
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	arr := strings.Split(header, " ")
	tokenValue := arr[1]
	if tokenValue != "" && tokenValue[:5] == "dummy" {
		err = user.DeleteDummyToken(arr[1])
		if err != nil {
			return
		}
	}
	ctx.JSON(data)
}
