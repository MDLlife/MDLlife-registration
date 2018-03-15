package router

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/iris-contrib/middleware/cors"

	"../config"
	"../controller"
	"time"
	"github.com/kataras/iris/middleware/basicauth"
)

func Routes(app *iris.Application) {
	// use recover(y) middleware, to prevent crash all app on request
	app.Use(recover.New())

	if config.Config.Debug {
		crs := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
			Debug: true,
		})

		app.UseGlobal(crs)
	}

	captchaRoute := app.Party("/captcha")
	captchaRoute.Get("/id", controller.CaptchaId)
	captchaRoute.Get("/{captcha}", controller.CaptchaMedia)

	app.Get("/whitelist/confirm_email", controller.WhitelistConfirmEmail)
	app.Post("/whitelist/request", iris.LimitRequestBodySize(config.Config.MaxFileUploadSizeMb<<20), controller.WhitelistRequest)

	// admin section
	authConfig := basicauth.Config{
		Users:   map[string]string{config.Config.AdminLogin: config.Config.AdminPassword},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(1) * time.Minute,
	}

	authentication := basicauth.New(authConfig)

	admin := app.Party("/admin", authentication).AllowMethods(iris.MethodOptions)  // <- important for the preflight.
	{
		admin.Get("/basic-auth", func(ctx iris.Context) {}) // to check auth
		admin.Get("/whitelist/list", func(ctx iris.Context) {
			ctx.Text("admin hello!")
		})
	}
}
