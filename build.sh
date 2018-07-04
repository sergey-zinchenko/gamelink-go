#!/bin/sh
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
docker build -t registry.gitlab.smedialink.com/z/gamelink-go .
docker push registry.gitlab.smedialink.com/z/gamelink-go
