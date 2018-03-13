package main

import (
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"time"
	"regexp"
	"errors"
	"path/filepath"
	"strings"
	"mime/multipart"
	"database/sql"

	"gopkg.in/yaml.v2"

	"github.com/kataras/iris"

	"./ses"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"github.com/go-xorm/xorm"
	"github.com/lib/pq"
	"database/sql/driver"

	"github.com/dchest/captcha"
	"path"
	"bytes"
	"github.com/go-xorm/core"
)

/*
	go get ./...
*/

// VerificationStage Type enumeration
type VerificationStage uint8

const (
	EMAIL_NOT_VERIFIED VerificationStage = iota
	EMAIL_VERIFIED
	FIRST_STAGE
)

func (u *VerificationStage) Scan(value interface{}) error { *u = VerificationStage(value.(uint8)); return nil }
func (u VerificationStage) Value() (driver.Value, error)  { return uint8(u), nil }

// Whitelist is whitelist table structure.
type Whitelist struct {
	Id                int64
	PassportId        int64             `xorm:"not null unique"`
	SelfieId          sql.NullInt64
	Name              string            `xorm:"varchar(255) not null"`
	Email             string            `xorm:"varchar(255) not null unique"`
	Phone             string            `xorm:"varchar(255) not null"`
	Birthday          string            `xorm:"varchar(255) not null"`
	Country           string            `xorm:"varchar(255) not null"`
	VerificationStage VerificationStage `xorm:"not null default 0"`
	CreatedAt         time.Time         `xorm:"created"`
	UpdatedAt         time.Time         `xorm:"updated"`
}

func (w *Whitelist) TableName() string {
	return "whitelists"
}

// Photo is photo table structure.
type Photo struct {
	Id        int64
	Path      string    `xorm:"varchar(255) not null unique"`
	Extension string    `xorm:"varchar(5) not null"`
	CreatedAt time.Time `xorm:"created"`
}

func (p *Photo) TableName() string {
	return "photos"
}

type WhitelistToken struct {
	WhitelistId int64
	Token       string    `xorm:"varchar(128) not null pk"`
	CreatedAt   time.Time `xorm:"created"`
	ExpiredAt   time.Time
	UsedAt      pq.NullTime
}

func (wt *WhitelistToken) TableName() string {
	return "whitelist_tokens"
}

// Config file structure
type Config struct {
	Debug bool `yaml:"Debug"`

	AwsKey    string `yaml:"AwsKey"`
	AwsSecret string `yaml:"AwsSecret"`
	AwsRegion string `yaml:"AwsRegion"`

	NoReplyEmail string `yaml:"NoReplyEmail"`
	ReplyEmail   string `yaml:"ReplyEmail"`

	DatabaseDriver string `yaml:"DatabaseDriver"`
	DatabaseDSN    string `yaml:"DatabaseDSN"`

	MaxFileUploadSizeMb int64 `yaml:"MaxFileUploadSizeMb"`

	Port string `yaml:"Port"`
}

var (
	config = &Config{}

	nameValidatorRegex = regexp.MustCompile("(?:(\\pL|[-])+((?:\\s)+)?)")
	// YYYY-MM-DD // YYYY >= 1000 matches correct dates in months
	dateValidatorRegex = regexp.MustCompile("^(?:[1-9]\\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\\d|2[0-9])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31))$")
	// phone number, simple
	phoneNumberValidatorRegex = regexp.MustCompile("[0-9()\\pL\\s-+#]+")

	tokenRegex = regexp.MustCompile("[0-9a-zA-Z]+")
)

func init() {
	loadConfig()

	// Amazon SES setup
	ses.SetConfiguration(config.AwsKey, config.AwsSecret, config.AwsRegion)
}

func main() {
	app := iris.New()

	// load templates
	app.RegisterView(iris.HTML("./templates", ".html").Reload(!config.Debug))

	db, err := xorm.NewEngine(config.DatabaseDriver, config.DatabaseDSN)

	if err != nil {
		app.Logger().Fatalf("db failed to initialized: %v", err)
	}

	if config.Debug {
		db.ShowSQL(true) // Show SQL statement on standard output;
		db.Logger().SetLevel(core.LOG_DEBUG)
	}
	//db.SetMaxOpenConns(60)
	//db.SetMaxIdleConns(5)

	iris.RegisterOnInterrupt(func() {
		db.Close()
	})

	db.Sync2(new(Whitelist), new(Photo), new(WhitelistToken))

	routes(app, db)

	app.Run(iris.Addr(config.Port),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithPostMaxMemory(config.MaxFileUploadSizeMb<<20))
}

func routes(app *iris.Application, db *xorm.Engine) {
	captchaRoute := app.Party("/captcha")
	captchaRoute.Get("/id", func(ctx iris.Context) {
		ctx.Text(captcha.New())
	})
	captchaRoute.Get("/{captcha}", func(ctx iris.Context) {
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
	})

	app.Get("/whitelist/confirm_email", func(ctx iris.Context) {
		token := ctx.FormValue("token")

		if !tokenRegex.MatchString(token) {
			ctx.ViewData("message", "Invalid token data")
			ctx.View("email-confirmation-error.html")
			return
		}

		whitelistToken := &WhitelistToken{}
		has, err := db.Where("token = ? AND used_at IS NULL AND expired_at > ?", token, time.Now()).Get(whitelistToken)
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

		whitelist, err := whitelistToken.TokenConfirmed(db)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't confirm token in database. " + err.Error())
			return
		}

		ctx.ViewData("email", whitelist.Email)
		ctx.View("email-confirmed.html")
	})

	app.Post("/whitelist/request", iris.LimitRequestBodySize(config.MaxFileUploadSizeMb<<20), func(ctx iris.Context) {
		birthday := ctx.FormValue("birthday")
		if birthday == "" {
			birthday = combineDatetime(ctx.FormValue("year"), ctx.FormValue("month"), ctx.FormValue("day"))
		}

		whitelist := &Whitelist{
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
				errs[name] = value
			}
		}
		if passportErr != nil {
			errs["passport"] = errors.New(fmt.Sprintf("Filesize is very large. Allowed up to %v Mb", config.MaxFileUploadSizeMb))
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

		has, err := db.Where("email = ?", whitelist.Email).Exist(&Whitelist{})
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

		filePath, fileExt, err := saveFile(passportFile, passportInfo)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't save passport!\n\t" + err.Error())
			return
		}

		photo := &Photo{
			Path:      filePath,
			Extension: fileExt,
		}

		if _, err := db.InsertOne(photo); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't insert passport in database " + err.Error())
			return
		}

		whitelist.PassportId = photo.Id

		// Get selfie file from the request.
		selfieFile, selfieInfo, selfieErr := ctx.FormFile("selfie")
		if selfieErr == nil {
			filePath, fileExt, err := saveFile(selfieFile, selfieInfo)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				println("Can't save selfie!\n\t" + err.Error())
				return
			}

			selfie := &Photo{
				Path:      filePath,
				Extension: fileExt,
			}

			if _, err := db.InsertOne(selfie); err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				println("Can't insert selfie in database " + err.Error())
				return
			}

			whitelist.SelfieId = sql.NullInt64{Int64: selfie.Id, Valid: true}
		}

		token, err := whitelist.StoreData(db)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't insert whitelist " + err.Error())
			return
		}

		sendEmail(whitelist.Email, token)

		ctx.JSON(map[string]bool{"success": true})
	})
}

func sendEmail(to string, token string) {
	emailData := ses.Email{
		To:   to,
		From: config.NoReplyEmail,
		Text: "Your whitelist submission is well received.\n\n" +
			"To finish the whitelist application process please confirm your email by following the link/n" +
			"https://mdl.life/whitelist/confirm_email?token=" + token + "\n\n" +
			"The instructions of how to purchase the MDL Tokens to be send soon is confirmation that you have passed the whitelist.\n\n" +
			"For inquiries and support please contact support@mdl.life",
		HTML: "<h3 style=\"color:purple;\">Your whitelist submission is well received.</h3><br>" +
			"To finish the whitelist application process please confirm your email by clicking the link<br>" +
			"<a href=\"https://mdl.life/whitelist/confirm_email?token=" + token + "\">" + "https://mdl.life/whitelist/confirm_email?token=" + token + "</a><br><br>" +
			"The instructions of how to purchase the MDL Tokens to be send soon is confirmation that you have passed the whitelist.<br><br>" +
			"For inquiries and support please contact <a href=\"mailto:support@mdl.life\">support@mdl.life</a>",
		Subject: "MDL Talent Hub: Whitelist application received",
		ReplyTo: config.ReplyEmail,
	}

	ses.SendEmail(emailData)
}

func combineDatetime(y string, m string, d string) string {
	str := y + "-"

	if len(m) == 1 {
		str += "0"
	}

	str += m + "-"

	if len(d) == 1 {
		str += "0"
	}

	str += d

	return str
}

func loadConfig() {
	// which will try to find the 'filename' from current working dir too.
	yamlAbsPath, err := filepath.Abs("config.yml")
	if err != nil {
		println("Can't find example.config.yml " + err.Error())
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		println("Can't read example.config.yml " + err.Error())
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func (w Whitelist) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Name, validation.Required, validation.Match(nameValidatorRegex)),
		validation.Field(&w.Email, validation.Required, is.Email),
		validation.Field(&w.Phone, validation.Match(phoneNumberValidatorRegex)),
		validation.Field(&w.Birthday, validation.Required, validation.Match(dateValidatorRegex)),
		validation.Field(&w.Country, validation.Required, validation.Match(nameValidatorRegex)),
	)
}

func saveFile(file multipart.File, fileInfo *multipart.FileHeader) (path string, ext string, err error) {
	defer file.Close()
	ext = filepath.Ext(fileInfo.Filename)
	ext = strings.ToLower(ext[1:]) // remove dot and cast to lower case
	var (
		filename string
		imgPath  string
	)

	// generate a new name if file exists
	for {
		filename = RandomString(48)
		imgPath = filename[0:3] + "/" + filename[3:6]

		// create path / bug with 0644
		if err := os.MkdirAll("./uploads/"+imgPath, 0744); err != nil {
			return "", "", errors.New("Can't create image path: " + err.Error())
		}

		path = "./uploads/" + imgPath + "/" + filename + "." + ext
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
	}

	// Create a file
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0744)

	if err != nil {
		return "", "", errors.New("Can't save image: " + err.Error())
	}
	defer out.Close()

	io.Copy(out, file)

	return path, ext, nil
}

func (w *Whitelist) StoreData(db *xorm.Engine) (emailToken string, err error) {
	tx := db.NewSession()
	defer tx.Close()

	if err = tx.Begin(); err != nil {
		return "", err
	}

	if _, err = tx.InsertOne(w); err != nil {
		return "", err
	}

	var token string
	// regenerate if not unique
	for {
		token = SecureRandomString(35)
		has, err := tx.Where("whitelist_id = ? AND token = ?", w.Id, token).Exist(&WhitelistToken{})
		if err != nil {
			return "", err
		}
		if !has {
			break
		}
	}

	wt := &WhitelistToken{
		WhitelistId: w.Id,
		Token:       token,
		ExpiredAt:   time.Now().AddDate(0, 0, 7),
	}

	if _, err = tx.InsertOne(wt); err != nil {
		return "", err
	}

	return token, tx.Commit()
}

func (wt *WhitelistToken) TokenConfirmed(db *xorm.Engine) (w *Whitelist, err error) {
	tx := db.NewSession()
	defer tx.Close()

	if err = tx.Begin(); err != nil {
		return nil, err
	}

	wt.UsedAt = pq.NullTime{Time: time.Now(), Valid: true}
	if _, err = tx.ID(wt.Token).Cols("used_at").Update(wt); err != nil {
		return nil, err
	}

	w = &Whitelist{}
	has, err := tx.ID(wt.WhitelistId).Get(w)
	if !has {
		return nil, errors.New(fmt.Sprintf("Can't find whitelist with id: %v", wt.WhitelistId))
	}
	if err != nil {
		return nil, err
	}

	w.VerificationStage = EMAIL_VERIFIED
	if _, err = tx.ID(w.Id).Cols("verification_stage").Update(w); err != nil {
		return nil, err
	}

	return w, tx.Commit()
}
