package social

import (
	"time"
	"net/http"
)

var client *http.Client

func init() {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSHandshakeTimeout: 10 * time.Second,

	}
	client = &http.Client{Transport: tr}
}