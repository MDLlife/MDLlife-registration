package tests

import (
	"encoding/json"
	"testing"
	"github.com/kataras/iris/httptest"
	"github.com/iris-contrib/httpexpect"

	"../app"
	"../config"
)

func InitTestServer(t *testing.T) *httpexpect.Expect {
	config.Config.DatabaseDriver = "sqlite3"
	config.Config.DatabaseDSN = "./test.db"

	app := app.NewApp()

	return httptest.New(t, app)
}

func JsonObjectFromString(str string, t *testing.T) interface{} {
	var dat interface{}

	if err := json.Unmarshal([]byte(str), &dat); err != nil {
		t.Logf("Wrong json data: %v\n", err.Error())
	}

	return dat
}
