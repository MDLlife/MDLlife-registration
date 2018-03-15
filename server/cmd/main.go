package main

import (
	"../config"
	"../db"
	"../model"
	"../router"

	"github.com/kataras/iris"
)

func main() {
	app := iris.New()
	if config.Config.Debug {
		app.Logger().SetLevel("debug")
	}

	// load templates
	app.RegisterView(iris.HTML("./templates", ".html").Reload(!config.Config.Debug))

	engine, err := db.Init()
	if err != nil {
		app.Logger().Fatalf("db failed to initialized: %v", err)
	}

	iris.RegisterOnInterrupt(func() {
		engine.Close()
	})

	engine.Sync2(new(model.Whitelist), new(model.Photo), new(model.WhitelistToken))

	router.Routes(app)

	app.Run(iris.Addr(config.Config.Port),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithPostMaxMemory(config.Config.MaxFileUploadSizeMb<<20))
}
