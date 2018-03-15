package router

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/iris-contrib/middleware/cors"

	"../config"
	"../controller"
)

func Routes(app *iris.Application) {
	// use recover(y) middleware, to prevent crash all app on request
	app.Use(recover.New())

	if config.Config.Debug {
		crs := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
			AllowCredentials: true,
		})

		app.UseGlobal(crs)
	}

	captchaRoute := app.Party("/captcha")
	captchaRoute.Get("/id", controller.CaptchaId)
	captchaRoute.Get("/{captcha}", controller.CaptchaMedia)

	app.Get("/whitelist/confirm_email", controller.WhitelistConfirmEmail)
	app.Post("/whitelist/request", iris.LimitRequestBodySize(config.Config.MaxFileUploadSizeMb<<20), controller.WhitelistRequest)
}
