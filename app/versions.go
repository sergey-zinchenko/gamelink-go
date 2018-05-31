package app

import (
	"encoding/json"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"log"
	"net/http"
)

var (
	// BuildTime is a time label of the moment when the binary was built
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built
	Commit = "unset"
	// Release is a semantic version of current build
	Release = "unset"
)

//getAppInfo - write info about app
func (a *App) getAppInfo(ctx iris.Context) {

	info := struct {
		BuildTime string `json:"buildTime"`
		Commit    string `json:"commit"`
		Release   string `json:"release"`
	}{
		BuildTime, Commit, Release,
	}

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("Could not encode info data: %v", err)
		ctx.StatusCode(http.StatusServiceUnavailable)
		return
	}
	ctx.ContentType(context.ContentJSONHeaderValue)
	ctx.Write(body)
}
