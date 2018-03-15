package main

import (
	"../app"
	"../config"

	"github.com/kataras/iris"
)

func main() {
	app.NewApp().Run(iris.Addr(config.Config.Port),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithPostMaxMemory(config.Config.MaxFileUploadSizeMb<<20))
}
