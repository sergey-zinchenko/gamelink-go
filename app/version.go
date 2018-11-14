package app

import (
	"encoding/json"
	"gamelink-go/version"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"log"
)

func (a *App) version(ctx iris.Context) {
	info := struct {
		BuildTime string `json:"buildTime"`
		Commit    string `json:"commit"`
		Release   string `json:"release"`
	}{
		version.BuildTime, version.Commit, version.Release,
	}

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("Could not encode info data: %v", err)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.WriteString(string(body))
}
