package router

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/iris-contrib/middleware/cors"

	"../config"
	"../controller"
	controller_admin "../controller/admin"
	"time"
	"github.com/kataras/iris/middleware/basicauth"
)

func Routes(app *iris.Application) {
	// use recover(y) middleware, to prevent crash all app on request
	app.Use(recover.New())

	var origins []string = []string{"*"}

	crs := cors.New(cors.Options{
		AllowedOrigins:   origins, // allows everything, use that to change the hosts.
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		// Debug: true,
	})

	root := app.Party("/", crs).AllowMethods(iris.MethodOptions) // <- important for the preflight.

	captchaRoute := root.Party("/captcha")
	captchaRoute.Get("/id", controller.CaptchaId)
	captchaRoute.Get("/{captcha}", controller.CaptchaMedia)

	root.Get("/whitelist/confirm_email", controller.WhitelistConfirmEmail)
	root.Post("/whitelist/request", iris.LimitRequestBodySize(config.Config.MaxFileUploadSizeMb<<20), controller.WhitelistRequest)

	// admin section
	authConfig := basicauth.Config{
		Users:   map[string]string{config.Config.AdminLogin: config.Config.AdminPassword},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(1) * time.Minute,
	}

	authentication := basicauth.New(authConfig)

	admin := root.Party("/admin", authentication)
	{
		admin.Get("/basic-auth", func(ctx iris.Context) {}) // to check auth
		admin.Get("/whitelist/list", controller_admin.GetWhitelistList)
		admin.Post("/whitelist/accept/{id:int min(1)}", controller_admin.WhitelistAccept)
		admin.Post("/whitelist/decline/{id:int min(1)}", controller_admin.WhitelistDecline)
		admin.Post("/whitelist/question/{id:int min(1)}", controller_admin.WhitelistQuestion)
	}
}
