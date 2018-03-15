package controller

import (
	"strings"
	"errors"
	"fmt"

	"github.com/kataras/iris"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/dchest/captcha"
	"database/sql"

	"../config"
	"../utils"
	"../model"
	"../model/validation_rules"
	"../email"
)

func WhitelistRequest(ctx iris.Context) {
	birthday := ctx.FormValue("birthday")
	if birthday == "" {
		birthday = utils.CombineDatetime(ctx.FormValue("year"), ctx.FormValue("month"), ctx.FormValue("day"))
	}

	whitelist := &model.Whitelist{
		Name:     ctx.FormValue("name"),
		Email:    ctx.FormValue("email"),
		Country:  ctx.FormValue("country"),
		Birthday: birthday,
	}

	// Get the file from the request.
	passportFile, passportInfo, passportErr := ctx.FormFile("passport")

	var errs = validation.Errors{}

	if e, ok := whitelist.Validate().(validation.Errors); ok {
		for name, value := range e {
			errs[strings.ToLower(name[:1]) + name[1:]] = value
		}
	}
	if passportErr != nil {
		errs["passport"] = errors.New(fmt.Sprintf("Filesize is very large. Allowed up to %v Mb", config.Config.MaxFileUploadSizeMb))
	} else
	if passportInfo.Filename == "" {
		errs["passport"] = errors.New("Add a file of your passport")
	}
	if !captcha.VerifyString(ctx.FormValue("captchaId"), ctx.FormValue("captchaSolution")) {
		errs["captchaSolution"] = errors.New("Captcha check has been failed")
	}

	if len(errs) > 0 {
		ctx.StatusCode(iris.StatusUnprocessableEntity)
		ctx.JSON(map[string]interface{}{"errors": errs})
		return
	}

	has, err := whitelist.EmailExist()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		println("Can't find whitelist record in database.\n\t" + err.Error())
		return
	}
	if has {
		ctx.StatusCode(iris.StatusUnprocessableEntity)
		ctx.JSON(map[string]interface{}{"errors": map[string]string{"email": "This email is already registered."}})
		return
	}

	photo := &model.Photo{}
	if err := photo.StoreFile(passportFile, passportInfo); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		println("Can't save passport!\n\t" + err.Error())
		return
	}

	whitelist.PassportId = photo.Id

	// Get selfie file from the request.
	selfieFile, selfieInfo, selfieErr := ctx.FormFile("selfie")
	if selfieErr == nil {
		selfie := &model.Photo{}
		if err := selfie.StoreFile(selfieFile, selfieInfo); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't save selfie!\n\t" + err.Error())
			return
		}

		whitelist.SelfieId = sql.NullInt64{Int64: selfie.Id, Valid: true}
	}

	token, err := whitelist.StoreData()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		println("Can't insert whitelist " + err.Error())
		return
	}

	email.ConfirmEmail(whitelist.Email, token)

	ctx.JSON(map[string]bool{"success": true})
}

func WhitelistConfirmEmail(ctx iris.Context) {
	token := ctx.FormValue("token")

	if !validation_rules.TokenRegex.MatchString(token) {
		ctx.ViewData("message", "Invalid token data")
		ctx.View("email-confirmation-error.html")
		return
	}

	whitelistToken := &model.WhitelistToken{}
	has, err := whitelistToken.GetValidToken(token)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		println("Token database search error. " + err.Error())
		return
	}
	if !has {
		ctx.ViewData("message", "You can only activate your email once.")
		ctx.View("email-confirmation-error.html")
		return
	}

	whitelist, err := whitelistToken.TokenConfirmed()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		println("Can't confirm token in database. " + err.Error())
		return
	}

	ctx.ViewData("email", whitelist.Email)
	ctx.View("email-confirmed.html")
}
