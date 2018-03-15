package tests

import (
	"github.com/kataras/iris/httptest"
	"testing"

	"../config"
)

func TestAdminAuth(t *testing.T) {
	e := InitTestServer(t)

	// redirects to /admin without basic auth
	e.GET("/admin/whitelist/list").Expect().Status(httptest.StatusUnauthorized)

	// with valid basic auth
	e.GET("/admin/whitelist/list").WithBasicAuth(config.Config.AdminLogin, config.Config.AdminPassword).Expect().
		Status(httptest.StatusOK).Body()

	// with invalid basic auth
	e.GET("/admin/whitelist/list").WithBasicAuth("invalidusername", "invalidpassword").
		Expect().Status(httptest.StatusUnauthorized)
}
