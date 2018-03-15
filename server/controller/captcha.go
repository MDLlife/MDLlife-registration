package controller

import (
	"github.com/kataras/iris"
	"github.com/dchest/captcha"
	"path"
	"strings"
	"bytes"
	"time"
)

func CaptchaId(ctx iris.Context) {
	ctx.Text(captcha.New())
}

func CaptchaMedia(ctx iris.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	dir, file := path.Split(ctx.Params().Get("captcha"))
	ext := path.Ext(file)
	id := file[:len(file)-len(ext)]
	if ext == "" || id == "" {
		ctx.StatusCode(iris.StatusNotFound)
		return
	}
	if ctx.FormValue("reload") != "" {
		captcha.Reload(id)
	}
	lang := strings.ToLower(ctx.FormValue("lang"))
	download := path.Base(dir) == "download"

	// send bytes buffer instead file
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")

	var content bytes.Buffer
	switch ext {
	case ".png":
		ctx.Header("Content-Type", "image/png")
		captcha.WriteImage(&content, id, captcha.StdWidth, captcha.StdHeight)
	case ".wav":
		ctx.Header("Content-Type", "audio/x-wav")
		captcha.WriteAudio(&content, id, lang)
	default:
		ctx.StatusCode(iris.StatusNotFound)
		return
	}

	if download {
		ctx.Header("Content-Type", "application/octet-stream")
	}

	ctx.ServeContent(bytes.NewReader(content.Bytes()), id+ext, time.Time{}, true)
}
