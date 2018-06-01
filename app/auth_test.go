package app

import (
	"gamelink-go/graceful"
	"github.com/go-redis/redis"
	"github.com/kataras/iris/httptest"
	"testing"
)

func TestAuthorizationMiddleware(t *testing.T) {
	checkAuthToken = func(token string, rc *redis.Client) (int64, *graceful.Error) {
		switch token {
		case "success":
			return 1, nil
		case "fatal":
			return 0, graceful.NewInvalidError("fatal test data")
		default:
			return 0, graceful.NewNotFoundError("wrong test data")
		}
	}
	t.Log("initializing server")
	app := NewApp()
	e := httptest.New(t, app.iris)
	t.Log("auth middleware test")
	e.GET("/users").Expect().Status(httptest.StatusUnauthorized)
	e.GET("/users").WithHeader("Authorization", "fail").Expect().Status(httptest.StatusUnauthorized)
	e.GET("/users").WithHeader("Authorization", "fatal").Expect().Status(httptest.StatusInternalServerError)
	e.GET("/users").WithHeader("Authorization", "success").Expect().Status(httptest.StatusOK).JSON().Path("$.userID").Equal(1)
	t.Log("auth middleware test ok")
}
